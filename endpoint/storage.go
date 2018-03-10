// microservices project main.go
package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	pb "github.com/olesho/spate/endpoint/proto"
)

type Storage struct {
	db *gorm.DB
}

type StorageConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewStorage(conf *StorageConfig) (*Storage, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8&parseTime=True&loc=Local", conf.User, conf.Password, conf.Host, conf.Port, conf.DBName))
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&pb.Endpoint{})
	db.Model(&pb.Endpoint{}).ModifyColumn("body", "text").ModifyColumn("header", "text").AddUniqueIndex("idx_url_user", "url", "user")
	return &Storage{db}, nil
}

func (s *Storage) Create(e *pb.Endpoint) (*pb.Endpoint, error) {
	err := s.db.Create(e).Error
	if err != nil {
		return nil, err
	}
	return e, nil
}
func (s *Storage) Read(eid *pb.EndpointID) (*pb.Endpoint, error) {
	e := &pb.Endpoint{}
	err := s.db.Where("id = ?", eid.Id).First(e).Error
	return e, err
}
func (s *Storage) Update(e *pb.Endpoint) (*pb.Endpoint, error) {
	err := s.db.Where("id = ?", e.ID).Update(e).Error
	return e, err
}
func (s *Storage) Delete(eid *pb.EndpointID) error {
	return s.db.Delete("id = ?", eid.Id).Error
}
func (s *Storage) List(uid *pb.UserID) ([]*pb.Endpoint, error) {
	endpoints := make([]*pb.Endpoint, 0)
	err := s.db.Where("user = ?", uid.Id).Find(&endpoints).Error
	return endpoints, err
}

func (s *Storage) Close() error {
	return s.db.Close()
}
