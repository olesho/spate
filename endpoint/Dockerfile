FROM golang:latest as builder
RUN go get github.com/olesho/spate/models
WORKDIR /go/src/github.com/olesho/spate/endpoint
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/olesho/spate/endpoint .
EXPOSE 3000
CMD ["./endpoint"]

