# ISV Serverless Nodepool Tagger, Ed Shnekendorf, September 2021
# Copyright (c) 2021 Oracle and/or its affiliates.
FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /docker-os-tester

CMD [ "/docker-os-tester" ]