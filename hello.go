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
	http.HandleFunc("/login", LoginHandle)
	appengine.Main()
}

type Error struct{
	Message string `json:"message"`
	Type string `json:"type"`
}

type ActionResponse struct{
	Type string `json:"type"`
	Action string `json:"action"`
	EMail string `json:"email"`
	ID string `json:"id"`
	Admin bool `json:"admin"`
}

func errorResponse500(w http.ResponseWriter, e error){
	errorResponse(w, fmt.Sprintf("%v", e), http.StatusInternalServerError)
}

func errorResponse(w http.ResponseWriter, message string, code int){
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	encoder := json.NewEncoder(w)

	encoder.Encode(Error{
		Type: "error",
		Message: message})
}

func emailHandle(w http.ResponseWriter, r *http.Request){
	if r.Method != "POST" {
		errorResponse(w, "Only 'POST' requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := appengine.NewContext(r)

	u,err := TokenAuth(ctx, r)

	if u == nil {
		errorResponse(w, "login missing", http.StatusUnauthorized)
		return
	}

	if err != nil {
		errorResponse500(w, err)
		return
	}


	decoder := json.NewDecoder(r.Body)
	email := Email{}

	err = decoder.Decode(&email)

	if err != nil {
		errorResponse500(w, err)
		return
	}

	email.UserId = u.Id()

	_, err = datastore.Put(ctx, datastore.NewKey(ctx, "email", email.Id,0, nil), &email)

	if err != nil {
		errorResponse500(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)

	encoder.Encode(ActionResponse{
		Type: "action",
		Action: "do_nothing",
		EMail: u.Email(),
		ID: u.Id(),
		Admin: u.IsAdmin(),
	})
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
