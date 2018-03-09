// feedman project main.go
package main

import (
	"encoding/json"
	"io"

	"log"
	"net/http"
	"strconv"

	"github.com/olesho/auth"
	pbendpoint "github.com/olesho/spate/endpoint/proto"
	"golang.org/x/net/context"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

func extractUserEndpointId(r *http.Request) (int64, int64, error) {
	uidStr := auth.UserIDbyCtx(r.Context())
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	return uid, id, err
}

func getEndpoint(r *http.Request, w http.ResponseWriter) {
	var e *pbendpoint.Endpoint
	uid, id, err := extractUserEndpointId(r)
	if err != nil {
		endpointListResp, err := endpointClient.List(context.TODO(), &pbendpoint.UserID{uid})
		if err != nil {
			return
		}
		if endpointListResp.Response.Ok {
			if len(endpointListResp.List) > 0 {
				e = endpointListResp.List[0]
			}
		}
	} else {
		endpointResp, err := endpointClient.Read(context.TODO(), &pbendpoint.EndpointID{id})
		if err != nil {
			log.Printf("Error reading endpoint: %v", err)
			return
		}

		e = endpointResp.Endpoint
	}

	if e != nil {
		err := json.NewEncoder(w).Encode(e)
		if err != nil {
			log.Printf("Error encoding '%v' endpoint: %v", id, err)
		}
		return
	}

	w.WriteHeader(404)
	return
}

func postEndpoint(r *http.Request) error {
	uidStr := auth.UserIDbyCtx(r.Context())
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
	}

	var endpoint *pbendpoint.Endpoint
	err = json.NewDecoder(r.Body).Decode(&endpoint)
	if err != nil {
		log.Printf("Error decoding endpoint from POST: %v", err)
		return err
	}

	endpoint.User = uid
	_, err = endpointClient.Create(context.TODO(), endpoint)
	if err != nil {
		log.Printf("Error creating endoint: %v", err)
		return err
	}

	return nil
}

/*
func deleteEndpoint(r *http.Request, w http.ResponseWriter) {
	uid := UserID(auth.UserIDbyCtx(r.Context()))
	title := EndpointTitle(r.URL.Query().Get("title"))

	err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("endpoints")).Delete([]byte(title))
	})
	if err != nil {
		log.Printf("DELETE endpoint error: %v", err)
		w.WriteHeader(502)
		return
	}
	endpointStorage.Delete(uid, title)
	w.WriteHeader(200)
}
*/

func HandleEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		getEndpoint(r, w)
		return
	}
	if r.Method == "POST" || r.Method == "PUT" {
		defer r.Body.Close()
		err := postEndpoint(r)
		if err != nil {
			log.Printf("POST endpoint error: %v", err)
			log.Println(err)
			w.WriteHeader(502)
			return
		}
		w.WriteHeader(200)
		return
	}
	if r.Method == "DELETE" {
		//		deleteEndpoint(r, w)
		return
	}
}

type tuple struct {
	Key string
	Val string
}
