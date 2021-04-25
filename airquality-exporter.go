package main

import (
	"net/http"

	"gopkg.in/alecthomas/kingpin.v2"
	log "github.com/sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryszard/sds011/go/sds011"
)

const (
	ver string = "0.19"
	logDateLayout string = "2006-01-02 15:04:05"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9999").String()
	portPath = kingpin.Flag("port-path", "Serial port path.").Default("/dev/ttyUSB0").String()
	cycle = kingpin.Flag("cycle", "Sensor cycle length in minutes.").Default("5").Int()
	forceSetCycle = kingpin.Flag("force-set-cycle", "Force set cycle on every program start, avoids stalling sensor with no measurements.").Default("true").Bool()
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
)

var (
	airqualityPM = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "airquality_pm",
		Help: "Airquality PM metric",
	},
	[]string{"type"})
)

func recordMetrics() {
	sensor, err := sds011.New(*portPath)
	if err != nil {
		log.Fatalf("Cannot create sensor instance: %v", err)
	}
	defer sensor.Close()


	if err := sensor.MakePassive(); err != nil {
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

	log.Infof("Setting sensor cycle to %d minutes", *cycle)
	if err := sensor.SetCycle(uint8(*cycle)); err != nil {
		log.Fatalf("Cannot set current cycle: %v", err)
	}


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
