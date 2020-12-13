FROM python:3.9.0-alpine3.12

RUN pip3 install pyserial prometheus_client

# Library from https://github.com/ikalchev/py-sds011 (https://github.com/mrk-its/py-sds011)
COPY sds011 /root/sds011
COPY airquality-exporter.py /root/

CMD ["/root/airquality-exporter.py"]
