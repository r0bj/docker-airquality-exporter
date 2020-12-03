#!/usr/bin/env python

# Compatible with SDS011 sensor

from prometheus_client import start_http_server, Gauge
import serial
import time

pm25 = Gauge('airquality_pm25', 'Airquality PM 2.5 metric', ['type'])
pm10 = Gauge('airquality_pm10', 'Airquality PM 10 metric', ['type'])

ser = serial.Serial('/dev/ttyUSB0')

def process_request():
	data = []
	for index in range(0,10):
		datum = ser.read()
		data.append(datum)

	pmtwofive = int.from_bytes(b''.join(data[2:4]), byteorder='little') / 10
	pmten = int.from_bytes(b''.join(data[4:6]), byteorder='little') / 10

	pm25.labels('pm2.5').set(pmtwofive)
	pm10.labels('pm10').set(pmten)

	time.sleep(30)


if __name__ == '__main__':
	start_http_server(9999)
	while True:
		process_request()
