// microservices project main.go
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bluele/gcache"

	micro "github.com/micro/go-micro"
	pb "github.com/olesho/spate/models/endpoint"
)

type Handler struct {
	l       sync.Mutex
	running map[int64]chan struct{}
	mq      micro.Publisher
	cache   gcache.Cache
}

func NewHandler(mq micro.Publisher) *Handler {
	return &Handler{
		l:       sync.Mutex{},
		running: make(map[int64]chan struct{}),
		mq:      mq,
		cache:   gcache.New(20).LRU().Build(),
	}
}

func (h *Handler) Start(e *pb.Endpoint) error {
	h.l.Lock()
	defer h.l.Unlock()

	// Remove is already exists
	if _, ok := h.running[e.ID]; ok {
		h.running[e.ID] <- struct{}{}
		delete(h.running, e.ID)
	}

	ticker := time.NewTicker(time.Duration(e.MinInterval) * time.Millisecond)
	quit := make(chan struct{})
	h.running[e.ID] = quit
	go func() {
		for {
			select {
			case <-ticker.C:
				err := h.request(e)
				if err != nil {
					log.Printf("Error sending request to endpoint %v: %v", e.ID, err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (h *Handler) Stop(eid *pb.EndpointID) error {
	h.l.Lock()
	defer h.l.Unlock()
	if _, ok := h.running[eid.Id]; ok {
		h.running[eid.Id] <- struct{}{}
		delete(h.running, eid.Id)
	}
	return nil
}

func (h *Handler) Trigger(e *pb.Endpoint) error {
	return h.request(e)
}

func (h *Handler) Data(key *pb.Key) (*pb.Body, error) {
	data, err := h.cache.GetIFPresent(key.Key)
	if err != nil {
		return nil, err
	}
	if body, ok := data.(pb.Body); ok {
		return &body, nil
	}
	return nil, nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

func (h *Handler) request(e *pb.Endpoint) error {
	reader := bufio.NewReader(strings.NewReader(e.Header))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		//log.Printf("Error reading MIME header: %v", err)
	}

	u, err := url.Parse(e.Url)
	if err != nil {
		return err
	}
	r := http.Request{
		Method: e.Method,
		URL:    u,
		Header: http.Header(mimeHeader),
		Body:   nopCloser{bytes.NewBufferString(e.Body)},
	}

	resp, err := http.DefaultClient.Do(&r)
	if err != nil {
		return err
	}

	var rdr io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		rdr, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
	default:
		rdr = resp.Body
	}

	defer rdr.Close()
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}

	err = h.cache.Set(e.ID, pb.Body{
		Created: time.Now().String(),
		Data:    data,
	}) //, time.Duration(e.Interval.Min)*time.Millisecond)
	if err != nil {
		return err
	}

	return h.mq.Publish(context.Background(), &pb.DataEvent{
		User: e.User,
		Key:  e.ID,
		Url:  e.Url,
	})
}
