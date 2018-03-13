package main

import (
	"context"
	"log"

	micro "github.com/micro/go-micro"
	pbendpoint "github.com/olesho/spate/models/endpoint"
	pbsubscribe "github.com/olesho/spate/models/subscribe"
	"github.com/olesho/spate/processors/upwork"
)

func main() {
	srv := micro.NewService(
		// This name must match the package name given in your protobuf definition
		micro.Name("processor"),
		micro.Version("latest"),
	)
	srv.Init()

	subscribeClient := pbsubscribe.NewSubscribeServiceClient("subscribe", srv.Client())

	micro.RegisterSubscriber("endpoint.data", srv.Server(), &Subscriber{
		endpointClient: pbendpoint.NewEndpointServiceClient("endpoint", srv.Client()),
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
		err := p.Handle(event.User, event.Url, resp.Data)
		if err != nil {
			log.Printf("Error handling %v: %v", event.Url, err)
		}
	}

	return nil
}
