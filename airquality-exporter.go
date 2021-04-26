package main

import (
	"fmt"
	"net/http"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryszard/sds011/go/sds011"
)

const (
	ver string = "0.24"
	logDateLayout string = "2006-01-02 15:04:05"
	// 0 retries, exit on failure
	retries int = 0
	apiCallTimeout int = 10
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9999").String()
	portPath = kingpin.Flag("port-path", "Serial port path.").Default("/dev/ttyUSB0").String()
	cycle = kingpin.Flag("cycle", "Sensor cycle length in minutes.").Default("5").Int()
	forceSetCycle = kingpin.Flag("force-set-cycle", "Force set cycle on every program start.").Default("true").Bool()
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
)

var (
	airqualityPM = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "airquality_pm",
		Help: "Airquality PM metric",
	},
	[]string{"type"})
)

func sensorMakePassive(sensor *sds011.Sensor) error {
	var responseError error

	response := make(chan error)
	Loop:
		for retry := 0; retry <= retries; retry++ {
			if retry > 0 {
				log.Debugf("Retrying (%d) API call", retry)
				time.Sleep(time.Second * time.Duration(retry))
			}

			go func() {
				if err := sensor.MakePassive(); err == nil {
					response <- nil
				} else {
					log.Warnf("Cannot switch sensor to passive mode: %v", err)
					response <- fmt.Errorf("Cannot switch sensor to passive mode: %v", err)
				}
			}()

			select {
			case err := <-response:
				if err == nil {
					responseError = nil
					break Loop
				} else {
					responseError = err
					continue Loop
				}
			case <-time.After(time.Second * time.Duration(apiCallTimeout)):
				log.Warnf("Device API response timeout (%d retries)", retry)
				responseError = fmt.Errorf("Device API response timeout (%d retries)", retry)
				continue Loop
			}
		}

	if responseError != nil {
		return responseError
	}

	return nil
}

func recordMetrics() {
	sensor, err := sds011.New(*portPath)
	if err != nil {
		log.Fatalf("Cannot create sensor instance: %v", err)
	}
	defer sensor.Close()

	if err := sensorMakePassive(sensor); err != nil {
		log.Fatalf("Cannot switch sensor to passive mode: %v", err)
	}

	if *forceSetCycle {
		log.Infof("Setting sensor cycle to %d minutes", *cycle)
		if err := sensor.SetCycle(uint8(*cycle)); err != nil {
			log.Fatalf("Cannot set current cycle: %v", err)
		}
	} else {
		currentCycle, err := sensor.Cycle()
		if err != nil {
			log.Fatalf("Cannot get current cycle: %v", err)
		}
		if currentCycle != uint8(*cycle) {
			log.Infof("Setting sensor cycle to %d minutes", *cycle)
			if err := sensor.SetCycle(uint8(*cycle)); err != nil {
				log.Fatalf("Cannot set current cycle: %v", err)
			}
		}
	}

	log.Info("Switching sensor to active mode")
	if err := sensor.MakeActive(); err != nil {
		log.Fatalf("Cannot switch sensor to active mode: %v", err)
	}

	for {
		point, err := sensor.Get()
		if err != nil {
			log.Errorf("Getting sensor measurement error: %v", err)
			continue
		}

		log.Infof("Sensor measurement results: %s", point)
		airqualityPM.WithLabelValues("pm2.5").Set(point.PM25)
		airqualityPM.WithLabelValues("pm10").Set(point.PM10)
	}
}

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = logDateLayout
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true

	kingpin.Version(ver)
	kingpin.Parse()

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Infof("Starting, version %s", ver)

	go recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
