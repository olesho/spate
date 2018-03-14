FROM spate:latest as builder
WORKDIR /go/src/github.com/olesho/spate/cli
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/olesho/spate/cli .
EXPOSE 3000
CMD ["./cli"]
