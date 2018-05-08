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
	"google.golang.org/appengine/user"
	"context"
	"io"
)

type SiteHits struct{
	Path string
	Date time.Time
}

func main() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/email", emailHandle)
	http.HandleFunc("/login", LoginHandle)
	http.HandleFunc("/enc", encTest)
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
	New bool `json:"new"`
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

func messageWrite(ctx context.Context, userId string, email *Email) error {
	vault := NewVault(ctx, userId)

	writer, err := vault.EncryptTo(fmt.Sprintf("messages/%s/%s.gpg", userId, email.Id))

	// Do nothing on key missing
	if err == KeyMissing {
		return nil
	}

	if err != nil {
		return err
	}

	_, err = io.WriteString(writer, email.Body)

	if err != nil {
		writer.Close()
		return err
	}

	err = writer.Close()

	return err
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

	emailKey := datastore.NewKey(ctx, "email", email.Id,0,nil)
	count, err := datastore.NewQuery("email").Filter("__key__ =", emailKey).Count(ctx)

	if err != nil {
		errorResponse500(w, err)
		return
	}

	_, err = datastore.Put(ctx, datastore.NewKey(ctx, "email", email.Id,0, nil), &email)

	if err != nil {
		errorResponse500(w, err)
		return
	}

	err = messageWrite(ctx, u.Id(), &email)

	if err != nil {
		errorResponse500(w, err)
		return
	}

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)

	emailMatcher := HandleEmail(
		AwsPendingConsolidationsHandler,
		AwsSubscriptionNotificationHandler,
		AzureDailyReportHandler,
		CAWelcomeHandler)

	encoder.Encode(emailMatcher(&EmailRequest{
		Email:email,
		User:u,
		New: count == 0,
	}))
}

func encTest(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)

	if u == nil {
		url, _ := user.LoginURL(ctx, r.URL.Path)

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		return
	}

	vault := NewVault(ctx, u.ID)

	writer,err := vault.EncryptTo("test.gpg")

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}

	_, err = writer.Write([]byte("this is a secret message"))

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}

	err = writer.Close()

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}

	err = vault.ArmorPrint("test.gpg", w)

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
	}
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
