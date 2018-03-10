// microservices project main.go
package main

import (
	"context"
	"log"
	"os"

	micro "github.com/micro/go-micro"
	pb "github.com/olesho/spate/endpoint/proto"
)

func main() {

	storage, err := NewStorage(&StorageConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		DBName:   os.Getenv("MYSQL_DATABASE"),
	})
	if err != nil {
		log.Fatalf("failed to open storage: %v", err)
	}

	srv := micro.NewService(
		// This name must match the package name given in your protobuf definition
		micro.Name("go.micro.srv.endpoint"),
		micro.Version("latest"),
	)
	srv.Init()

	handler := NewHandler(micro.NewPublisher("endpoint.data", srv.Client()))
	serviceProvider := &service{
		handler: handler,
		storage: storage,
	}

	pb.RegisterEndpointServiceHandler(srv.Server(), serviceProvider)
	if err := srv.Run(); err != nil {
		log.Println(err)
	}
}

type EndpointHandler interface {
	Start(*pb.Endpoint) error
	Stop(*pb.EndpointID) error
	Trigger(*pb.Endpoint) error
	Data(*pb.Key) (*pb.Body, error)
}

type EndpointStorage interface {
	Create(*pb.Endpoint) (*pb.Endpoint, error)
	Read(*pb.EndpointID) (*pb.Endpoint, error)
	Update(*pb.Endpoint) (*pb.Endpoint, error)
	Delete(*pb.EndpointID) error
	List(*pb.UserID) ([]*pb.Endpoint, error)
}

type service struct {
	handler EndpointHandler
	storage EndpointStorage
}

func (s *service) Create(ctx context.Context, req *pb.Endpoint, res *pb.EndpointsResponse) error {
	endpoint, err := s.storage.Create(req)
	if err != nil {
		res.Response = &pb.Response{
			Ok:    false,
			Error: err.Error(),
		}
		return err
	}
	res.Response = &pb.Response{Ok: true}
	res.Endpoint = endpoint
	return err
}

func (s *service) Read(ctx context.Context, req *pb.EndpointID, res *pb.EndpointsResponse) error {
	endpoint, err := s.storage.Read(req)
	if err != nil {
		res.Response = &pb.Response{
			Ok:    false,
			Error: err.Error(),
		}
		return err
	}
	res.Response = &pb.Response{Ok: true}
	res.Endpoint = endpoint
	return err
}

func (s *service) Update(ctx context.Context, req *pb.Endpoint, res *pb.EndpointsResponse) error {
	endpoint, err := s.storage.Update(req)
	if err != nil {
		res.Response = &pb.Response{
			Ok:    false,
			Error: err.Error(),
		}
		return err
	}
	res.Response = &pb.Response{Ok: true}
	res.Endpoint = endpoint
	return nil
}

func (s *service) Delete(ctx context.Context, req *pb.EndpointID, res *pb.Response) error {
	err := s.storage.Delete(req)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}
	res.Ok = true
	return nil
}

func (s *service) List(ctx context.Context, req *pb.UserID, res *pb.EndpointsListResponse) error {
	list, err := s.storage.List(req)
	if err != nil {
		res.Response = &pb.Response{
			Ok:    false,
			Error: err.Error(),
		}
		return err
	}
	res.Response = &pb.Response{Ok: true}
	res.List = list
	return nil
}

func (s *service) Start(ctx context.Context, req *pb.EndpointID, res *pb.Response) error {
	endpoint, err := s.storage.Read(req)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}

	err = s.handler.Start(endpoint)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}
	res.Ok = true
	return nil
}

func (s *service) Stop(ctx context.Context, req *pb.EndpointID, res *pb.Response) error {
	err := s.handler.Stop(req)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}
	res.Ok = true
	return nil
}

func (s *service) Trigger(ctx context.Context, req *pb.EndpointID, res *pb.Response) error {
	endpoint, err := s.storage.Read(req)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}

	err = s.handler.Trigger(endpoint)
	if err != nil {
		res.Ok = false
		res.Error = err.Error()
		return err
	}
	res.Ok = true
	return nil
}

func (s *service) Data(ctx context.Context, req *pb.Key, res *pb.Body) error {
	data, err := s.handler.Data(req)
	if err != nil {
		return err
	}
	res.Data = data.Data
	res.Created = data.Created
	return nil
}
