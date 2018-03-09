package main

import (
	"context"
	"log"

	micro "github.com/micro/go-micro"
	pbendpoint "github.com/olesho/spate/endpoint/proto"
	"github.com/olesho/spate/processors/upwork"
	pbsubscribe "github.com/olesho/spate/subscribe/proto"
)

func main() {
	srv := micro.NewService(
		// This name must match the package name given in your protobuf definition
		micro.Name("go.micro.srv.processor"),
		micro.Version("latest"),
	)
	srv.Init()

	subscribeClient := pbsubscribe.NewSubscribeServiceClient("go.micro.srv.subscribe", srv.Client())

	micro.RegisterSubscriber("endpoint.data", srv.Server(), &Subscriber{
		endpointClient: pbendpoint.NewEndpointServiceClient("go.micro.srv.endpoint", srv.Client()),
		processors: []EndpointHandler{
			upwork.NewUpworkProcessor(subscribeClient),
		},
	})
	if err := srv.Run(); err != nil {
		log.Println(err)
	}

}

type EndpointHandler interface {
	Handle(user int64, url string, data []byte) error
}

type Subscriber struct {
	endpointClient pbendpoint.EndpointServiceClient
	processors     []EndpointHandler
}

func (sub *Subscriber) Process(ctx context.Context, event *pbendpoint.DataEvent) error {
	resp, err := sub.endpointClient.Data(context.TODO(), &pbendpoint.Key{event.Key})
	if err != nil {
		return err
	}

	for _, p := range sub.processors {
		p.Handle(event.User, event.Url, resp.Data)
	}

	return nil
}
