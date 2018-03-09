// feedman project main.go
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	pb "github.com/olesho/spate/subscribe/proto"
)

type FirebaseProvider struct {
	key           string
	subscriptions map[int64]*pb.Subscription
	l             sync.Mutex
}

func NewFirebaseProvider(key string) *FirebaseProvider {
	mp := &FirebaseProvider{
		key:           key,
		subscriptions: make(map[int64]*pb.Subscription),
		l:             sync.Mutex{},
	}
	return mp
}

type Notification struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	Icon        string `json:"icon"`
	ClickAction string `json:"click_action"`
}

type FirebaseNotification struct {
	Notification Notification `json:"notification"`
	To           string       `json:"to"`
}

func (mp *FirebaseProvider) Users() []int64 {
	res := []int64{}
	for uid, _ := range mp.subscriptions {
		res = append(res, uid)
	}
	return res
}

type Result struct {
	MessageID string `json:"message_id"`
}

type Results struct {
	Results []Result `json:"results"`
}

func (mp *FirebaseProvider) Notify(n *pb.Notification) error {
	if subscription, ok := mp.subscriptions[n.UserId]; ok {
		if subscription.Active {
			b, err := json.Marshal(&FirebaseNotification{
				Notification: Notification{
					Title:       n.Title,
					Body:        n.Body,
					ClickAction: n.Url,
				},
				To: subscription.Token,
			})
			if err != nil {
				return err
			}
			r, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(b))
			if err != nil {
				return err
			}

			r.Header.Add("Authorization", "key="+mp.key)
			r.Header.Add("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(r)
			defer resp.Body.Close()
			if err != nil {
				return err
			}

			var res interface{}
			err = json.NewDecoder(resp.Body).Decode(&res)
			if err != nil {
				return err
			}

			log.Printf("Response: %+v\n", res)

			return nil
		}
	}

	return ERR_NO_SUBSCRIPTION
}

func (mp *FirebaseProvider) Create(s *pb.Subscription) {
	mp.l.Lock()
	defer mp.l.Unlock()
	mp.subscriptions[s.UserId] = s
}

func (mp *FirebaseProvider) Delete(s *pb.User) {
	mp.l.Lock()
	defer mp.l.Unlock()
	delete(mp.subscriptions, s.UserId)
}

func (mp *FirebaseProvider) List() []*pb.Subscription {
	list := make([]*pb.Subscription, 0)
	for _, s := range mp.subscriptions {
		list = append(list, s)
	}
	return list
}

func (mp *FirebaseProvider) Start(s *pb.User) {
	mp.l.Lock()
	defer mp.l.Unlock()
	mp.subscriptions[s.UserId].Active = true
}

func (mp *FirebaseProvider) Stop(s *pb.User) {
	mp.l.Lock()
	defer mp.l.Unlock()
	mp.subscriptions[s.UserId].Active = false
}
