package main

import (
	"time"
	"google.golang.org/appengine/datastore"
	"encoding/json"
	"fmt"
)

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