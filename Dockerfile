FROM golang:latest as builder
RUN go get github.com/olesho/spate/models
WORKDIR /go/src/github.com/olesho/spate
RUN go get ./...
