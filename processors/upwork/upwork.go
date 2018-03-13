package upwork

import (
	"context"
	"encoding/xml"
	"log"
	"strings"

	strip "github.com/grokify/html-strip-tags-go"
	client "github.com/micro/go-micro/client"
	pb "github.com/olesho/spate/models/subscribe"
)

type Notifier interface {
	Notify(context.Context, *pb.Notification, ...client.CallOption) (*pb.Response, error)
}

type Upwork struct {
	items    []Item
	notifier Notifier
}

func NewUpworkProcessor(n Notifier) *Upwork {
	return &Upwork{
		items:    []Item{},
		notifier: n,
	}
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Rss struct {
	Channel Channel `xml:"channel"`
}

func (u *Upwork) match(url string) bool {
	return strings.HasPrefix(url, "https://www.upwork.com/ab/feed/topics/rss?securityToken=")
}

func (u *Upwork) Handle(uid int64, url string, data []byte) error {
	if u.match(url) {
		log.Printf("Handling %v", url)
		var feed Rss
		err := xml.Unmarshal(data, &feed)
		if err != nil {
			return err
		}

		if len(u.items) == 0 {
			u.items = feed.Channel.Items
		}

		for _, item := range feed.Channel.Items {
			if u.addUrl(&item) {
				resp, err := u.notifier.Notify(context.TODO(), &pb.Notification{
					Title:  item.Title,
					Body:   item.Description,
					Url:    item.Link,
					UserId: uid,
				})
				if err != nil {
					return err
				}
				if !resp.Ok {
					log.Printf("Unable to notify due to: %v", resp.Error)
				}
			}
		}
	}
	return nil
}

func (u *Upwork) addUrl(item *Item) bool {
	for _, found := range u.items {
		if item.Link == found.Link {
			return false
		}
	}
	item.Description = strip.StripTags(item.Description)
	u.items = append([]Item{*item}, u.items...)
	return true
}
