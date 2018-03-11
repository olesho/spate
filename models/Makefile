build:
	protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/olesho/spate/models \
	  endpoint/endpoint.proto
	protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/olesho/spate/models \
	  subscribe/subscribe.proto