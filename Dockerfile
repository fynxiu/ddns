FROM golang:1.13.1-alpine

ENV GOPROXY=https://goproxy.io

ADD . /etc/ddns

RUN cd /etc/ddns && go mod download && go build && mv ddns /usr/local/bin/

CMD ["ddns"]
