// microservices project main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/boltdb/bolt"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	pb "github.com/olesho/spate/models/subscribe"
)

var db *bolt.DB

func Advertise(a string) server.Option {
	return func(o *server.Options) {
		o.Advertise = a
	}
}

func main() {
	subscriptionProvider := NewFirebaseProvider(os.Getenv("GCM_API_KEY"))

	s := server.NewServer(
		server.Name("subscribe"),
		server.Version("latest"),
		Advertise(os.Getenv("ADVERTISE")),
	)

	srv := micro.NewService(
		micro.Server(s),
	)

	srv.Init()
	pb.RegisterSubscribeServiceHandler(srv.Server(), &service{subscriptionProvider})
	if err := srv.Run(); err != nil {
		log.Println(err)
	}
}

type Provider interface {
	Create(*pb.Subscription)
	Delete(*pb.User)
	List() []*pb.Subscription
	Notify(*pb.Notification) error
}

type service struct {
	provider Provider
}

func (s *service) Create(ctx context.Context, req *pb.Subscription, res *pb.Response) error {
	s.provider.Create(req)
	res.Ok = true
	return nil
}

func (s *service) Delete(ctx context.Context, req *pb.User, res *pb.Response) error {
	s.provider.Delete(req)
	res.Ok = true
	return nil
}

func (s *service) List(ctx context.Context, req *pb.EmptySubscription, res *pb.SubscriptionsList) error {
	l := s.provider.List()
	res.Response = &pb.Response{Ok: true}
	res.List = l
	return nil
}

func (s *service) Notify(ctx context.Context, req *pb.Notification, res *pb.Response) error {
	err := s.provider.Notify(req)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
	}
	res.Ok = true
	return nil
}

func (s *service) Status(ctx context.Context, req *pb.EmptySubscription, res *pb.Response) error {
	res.Ok = true
	return nil
}
