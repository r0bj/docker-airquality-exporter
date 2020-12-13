#!/usr/bin/env python

# Compatible with SDS011 sensor

from prometheus_client import start_http_server, Gauge
import serial
import time
import sds011 # From https://github.com/mrk-its/py-sds011

pm = Gauge('airquality_pm', 'Airquality PM metric', ['type'])

def process_request(sensor):
	sensor.sleep(sleep=False)
	time.sleep(30)
	results = sensor.query()
	sensor.sleep()

	pm.labels('pm2.5').set(results[0])
	pm.labels('pm10').set(results[1])

	time.sleep(360)

if __name__ == '__main__':
	start_http_server(9999)

	sensor = sds011.SDS011("/dev/ttyUSB0", use_query_mode=True)
	while True:
		process_request(sensor)
