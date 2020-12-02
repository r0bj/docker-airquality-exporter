FROM python:3.9.0-alpine3.12

RUN pip3 install pyserial prometheus_client

COPY airquality-exporter.py /
