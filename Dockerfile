FROM python:3.9.0-alpine3.12

RUN pip3 install pyserial prometheus_client https://github.com/mrk-its/py-sds011/archive/v0.9.zip#py-sds011==0.9

COPY airquality-exporter.py /

CMD ["/airquality-exporter.py"]
