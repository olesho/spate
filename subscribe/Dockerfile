FROM golang:latest
WORKDIR /go/src/github.com/olesho/spate/subscribe
COPY . .
RUN go get -u ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/olesho/spate/subscribe .
CMD ["./subscribe"]