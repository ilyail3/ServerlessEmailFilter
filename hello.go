// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"time"
	"google.golang.org/appengine/datastore"
	"encoding/json"
)

type SiteHits struct{
	Path string
	Date time.Time
}

func main() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/email", emailHandle)
	appengine.Main()
}

type EmailAddress struct {
	Address string `json:"address"`
	Name string `json:"name"`
}

type Email struct{
	Subject string `json:"subject"`
	Body string `json:"body"`
	From EmailAddress `json:"from"`
	ToAddress []EmailAddress `json:"toAddress"`
	CcAddress []EmailAddress `json:"ccAddress"`
	BccAddress []EmailAddress `json:"bccAddress"`
	SendTime time.Time `json:"sendTime"`
	ReceiveTime time.Time `json:"receiveTime"`
	CreateTime time.Time `json:"createTime"`
}

type Error struct{
	Message string `json:"message"`
	Type string `json:"type"`
}

type ActionResponse struct{
	Type string `json:"type"`
	Action string `json:"action"`
}

func emailHandle(w http.ResponseWriter, r *http.Request){
	encoder := json.NewEncoder(w)

	if r.Method != "POST" {
		encoder.Encode(Error{
			Type: "error",
			Message: "Only 'POST' requests are allowed"})

		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	email := Email{}

	err := decoder.Decode(&email)

	if err != nil {
		encoder.Encode(Error{
			Type: "error",
			Message: fmt.Sprintf("%v", err)})

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	encoder.Encode(ActionResponse{
		Type: "action",
		Action: "do_nothing"})
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	log.Infof(ctx,"test url:%v" , r.URL)

	hit := SiteHits{
		Path: r.URL.Path,
		Date: time.Now(),
	}

	key, err := datastore.Put(ctx, datastore.NewIncompleteKey(ctx, "site-hits", nil), &hit)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Infof(ctx, "Stored hit as %v", key)

	fmt.Fprintln(w, "Hello, world! 12")
}
