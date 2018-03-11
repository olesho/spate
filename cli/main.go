package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	micro "github.com/micro/go-micro"
	"github.com/olesho/auth"
	pbendpoint "github.com/olesho/spate/models/endpoint"
	pbsubscribe "github.com/olesho/spate/models/subscribe"
	"golang.org/x/net/context"
)

func parse(data string) (*pbsubscribe.Subscription, error) {
	var s *pbsubscribe.Subscription
	err := json.Unmarshal([]byte(data), &s)
	return s, err
}

var service micro.Service
var subscribeClient pbsubscribe.SubscribeServiceClient
var endpointClient pbendpoint.EndpointServiceClient

func init() {
	service = micro.NewService(micro.Name("feed.client"))
	service.Init()

	subscribeClient = pbsubscribe.NewSubscribeServiceClient("go.micro.srv.subscribe", service.Client())
	endpointClient = pbendpoint.NewEndpointServiceClient("go.micro.srv.endpoint", service.Client())
}

func main() {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	fbCallbackURL := os.Getenv("FB_CALLBACK_URL")
	secure, _ := strconv.ParseBool(os.Getenv("SECURE"))
	domain := os.Getenv("DOMAIN")
	p := auth.NewFacebookAuthProvider(auth.FacebookProviderConfig{
		ClientID:      os.Getenv("FB_ID"),
		ClientSecret:  os.Getenv("FB_SECRET"),
		SecureCookies: secure,
		Domain:        domain,
		CallbackURL:   fbCallbackURL,
		SuccessURL:    "/",
		FailURL:       "/auth/facebook",
	}, newStorage())
	http.HandleFunc("/auth/facebook", p.HandleFacebook)
	http.HandleFunc("/auth/facebook/callback", p.HandleFacebookCallback)
	http.HandleFunc("/subscribe/firebase", p.Middleware(func(w http.ResponseWriter, r *http.Request) {
		uidStr := auth.UserIDbyCtx(r.Context())
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing user ID: %v", err)
			w.WriteHeader(502)
			return
		}

		token := r.Header.Get("subscription")
		createSubscriptionResp, err := subscribeClient.Create(context.Background(), &pbsubscribe.Subscription{
			UserId: uid,
			Token:  token,
			Active: true,
		})
		if err != nil {
			log.Printf("Error subscribing client %v: %v", uid, err)
			w.WriteHeader(502)
			return
		}
		if !createSubscriptionResp.Ok {
			log.Printf("Error subscribing client %v: %v", uid, createSubscriptionResp.Error)
			w.WriteHeader(502)
		}
		w.WriteHeader(200)
	}))

	http.HandleFunc("/start", p.Middleware(func(w http.ResponseWriter, r *http.Request) {
		uid, id, err := extractUserEndpointId(r)
		if err != nil {
			respList, err := endpointClient.List(context.Background(), &pbendpoint.UserID{uid})
			if err != nil {
				log.Printf("Error getting endpoints list: %v", err)
				w.WriteHeader(404)
				return
			}

			if len(respList.List) > 0 {
				id = respList.List[0].ID
			}
		}

		triggerResp, err := endpointClient.Start(context.TODO(), &pbendpoint.EndpointID{id})
		if err != nil {
			log.Printf("Error starting endoint: %v", err)
			w.WriteHeader(502)
			return
		}

		if triggerResp.Ok {
			w.WriteHeader(200)
		}
	}))

	http.HandleFunc("/stop", p.Middleware(func(w http.ResponseWriter, r *http.Request) {
		uidStr := auth.UserIDbyCtx(r.Context())
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing user ID: %v", err)
		}
		respList, _ := endpointClient.List(context.Background(), &pbendpoint.UserID{uid})
		for _, s := range respList.List {
			log.Printf("Item: %+v\n", *s)
		}
		if respList.Response.Ok {
			triggerResp, err := endpointClient.Stop(context.TODO(), &pbendpoint.EndpointID{respList.List[0].ID})
			if err != nil {
				log.Printf("Error stopping endoint: %v", err)
				return
			}
			log.Printf("Stop resp %+v\n", triggerResp)
		}
	}))

	http.HandleFunc("/subscription/list", p.Middleware(func(w http.ResponseWriter, r *http.Request) {
		resp, err := subscribeClient.List(context.TODO(), &pbsubscribe.EmptySubscription{})
		if err != nil {
			log.Printf("Unable to get subscriptions list", err)
			w.WriteHeader(502)
			return
		}
		if !resp.Response.Ok {
			log.Printf("Unable to get subscriptions list", resp.Response.Error)
			w.WriteHeader(502)
			return
		}
		log.Printf("Subscriptions list: %+v\n", resp.List)
		w.WriteHeader(200)
	}))

	http.HandleFunc("/ping", p.Middleware(func(w http.ResponseWriter, r *http.Request) {
		uidStr := auth.UserIDbyCtx(r.Context())
		uid, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing user ID: %v", err)
		}

		resp, err := subscribeClient.Notify(context.TODO(), &pbsubscribe.Notification{
			UserId: uid,
			Title:  "Hello",
			Body:   "Thy shall not pass!",
			Url:    "https://google.com",
		})
		if err != nil {
			log.Printf("Unable to get subscriptions list", err)
			return
		}
		if !resp.Ok {
			log.Printf("Unable to get subscriptions list", resp.Error)
			return
		}

		w.Write([]byte("Notification sent successfully"))
	}))

	http.HandleFunc("/endpoint", p.Middleware(HandleEndpoint))
	http.Handle("/", p.MiddlewareHandler(&staticHandler{}))

	log.Printf("Listening to %v:%v", host, port)
	err := http.ListenAndServe(host+":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

type staticHandler struct {
}

func (sh *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.String() == "/" {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Expires", "0")
	}
	http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
}
