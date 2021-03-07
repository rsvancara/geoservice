 FROM golang:1.16.0-alpine3.13 as builder

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

# Set up build directories
RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

# Build the goblog binary
COPY cmd /BUILD/cmd
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY internal /BUILD/internal
RUN cd /BUILD && go build -o /BUILD/geoservice cmd/geolookup/main.go 

#
# Maxmind
#
FROM  debian:stretch-slim as maxmindupdate

#Should be set in environment
ARG ACCOUNT_ID="123"
ARG LICENSE_KEY="xxx"

RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

RUN echo $ACCOUNT_ID

RUN \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y wget ca-certificates && \
  apt-get clean

WORKDIR /BUILD

RUN wget https://github.com/maxmind/geoipupdate/releases/download/v4.6.0/geoipupdate_4.6.0_linux_amd64.deb && \
    dpkg -i geoipupdate_4.6.0_linux_amd64.deb 

RUN echo "AccountID ${ACCOUNT_ID}" > /etc/GeoIP.conf && \
    echo "LicenseKey ${LICENSE_KEY}" >> /etc/GeoIP.conf && \
    echo "EditionIDs GeoIP2-City GeoIP2-Country GeoLite2-ASN GeoLite2-City GeoLite2-Country" >> /etc/GeoIP.conf && \
    echo "DatabaseDirectory /BUILD/db" >> /etc/GeoIP.conf && \
    /usr/bin/geoipupdate -v

#
# Running container
#
FROM alpine:3.13
RUN apk add ca-certificates

# Add user and set up temporary account
RUN mkdir /app && \
    mkdir app/temp && \
    addgroup app && \
    addgroup geoservice && \
    adduser --home /app --system --no-create-home geoservice geoservice && \
    chown -R geoservice:geoservice /app && \
    chmod 1777 app/temp 

COPY --from=builder /BUILD/geoservice /app/geoservice
COPY --from=maxmindupdate /BUILD/db /app/db

WORKDIR /app

USER geoservice

# This container exposes port 4990 to the outside world
EXPOSE 4990

CMD ["./geoservice"]