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

type EmailAddress struct {
	Address string `json:"address"`
	Name string `json:"name"`
}

type Email struct{
	Id string `json:"id"`
	Subject string `json:"subject"`
	Body string `json:"body"`
	From EmailAddress `json:"from"`
	ToAddress []EmailAddress `json:"toAddress"`
	CcAddress []EmailAddress `json:"ccAddress"`
	BccAddress []EmailAddress `json:"bccAddress"`
	SendTime time.Time `json:"sendTime"`
	ReceiveTime time.Time `json:"receiveTime"`
	CreateTime time.Time `json:"createTime"`
	UserId string
}

func addressToJson(emails []EmailAddress, name string)([]datastore.Property,error){
	items := make([]datastore.Property, len(emails))

	for i, add := range emails {
		bytes, err := json.Marshal(add)

		if err != nil {
			return items, err
		}

		items[i].Name = name
		items[i].Value = string(bytes)
		items[i].Multiple = true
	}

	return items, nil
}

func jsonToAddress(address string)(EmailAddress, error){
	result := EmailAddress{}

	err := json.Unmarshal([]byte(address), &result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (e *Email)Save()([]datastore.Property, error){
	bytes, err := json.Marshal(e.From)

	if err != nil {
		return []datastore.Property{}, err
	}

	result := []datastore.Property{
		{Name:"subject", Value:e.Subject},
		{Name:"body", Value:e.Body},
		{Name:"from", Value:string(bytes)},
		{Name:"send_time", Value:e.SendTime},
		{Name:"receive_time", Value:e.ReceiveTime},
		{Name:"create_time", Value:e.CreateTime},
		{Name:"user_id", Value:e.UserId},
	}

	toAddress, err := addressToJson(e.ToAddress, "to_address")

	if err != nil {
		return []datastore.Property{}, err
	}

	result = append(result, toAddress...)

	ccAddress, err := addressToJson(e.CcAddress, "cc_address")

	if err != nil {
		return []datastore.Property{}, err
	}

	result = append(result, ccAddress...)

	bccAddress, err := addressToJson(e.BccAddress, "bcc_address")

	if err != nil {
		return []datastore.Property{}, err
	}

	result = append(result, bccAddress...)

	return result, nil
}

func (e *Email) Load(props []datastore.Property) error{
	var err error = nil
	var address EmailAddress

	e.ToAddress = make([]EmailAddress, 0)
	e.CcAddress = make([]EmailAddress, 0)
	e.BccAddress = make([]EmailAddress, 0)

	for _, p := range props{
		if p.Name == "subject" {
			e.Subject = p.Value.(string)
		} else if p.Name == "body" {
			e.Body = p.Value.(string)
		} else if p.Name == "to_address" {
			e.From,err = jsonToAddress(p.Value.(string))
		} else if p.Name == "send_time" {
			e.SendTime = p.Value.(time.Time)
		} else if p.Name == "receive_time" {
			e.ReceiveTime = p.Value.(time.Time)
		} else if p.Name == "create_time" {
			e.CreateTime = p.Value.(time.Time)
		} else if p.Name == "user_id" {
			e.UserId = p.Value.(string)
		} else if p.Name == "to_address" {
			address, err = jsonToAddress(p.Value.(string))
			e.ToAddress = append(e.ToAddress, address)
		} else if p.Name == "cc_address" {
			address, err = jsonToAddress(p.Value.(string))
			e.CcAddress = append(e.CcAddress, address)
		} else if p.Name == "bcc_address" {
			address, err = jsonToAddress(p.Value.(string))
			e.BccAddress = append(e.BccAddress, address)
		}

		if err != nil {
			return fmt.Errorf(
				"failed to process field:%s, error:%v",
				p.Name,
				err)
		}

		fmt.Printf("Property:%s Value:%#v\n", p.Name, p.Value)
	}

	return nil
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

func emailHandle(w http.ResponseWriter, r *http.Request){

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)

	if r.Method != "POST" {
		encoder.Encode(Error{
			Type: "error",
			Message: "Only 'POST' requests are allowed"})

		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := appengine.NewContext(r)

	u,err := TokenAuth(ctx, r)

	if u == nil {
		encoder.Encode(Error{
			Type: "error",
			Message: "login missing"})

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err != nil {
		encoder.Encode(Error{
			Type: "error",
			Message: fmt.Sprintf("%v", err)})

		w.WriteHeader(http.StatusInternalServerError)
		return
	}


	decoder := json.NewDecoder(r.Body)
	email := Email{}

	err = decoder.Decode(&email)

	if err != nil {
		encoder.Encode(Error{
			Type: "error",
			Message: fmt.Sprintf("%v", err)})

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	email.UserId = u.Id()

	_, err = datastore.Put(ctx, datastore.NewKey(ctx, "email", email.Id,0, nil), &email)

	if err != nil {
		encoder.Encode(Error{
			Type: "error",
			Message: fmt.Sprintf("%v", err)})

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
