package main

import (
	"testing"
	"context"
	"github.com/ilyail3/ServerlessEmailFilter/comp"
)

type TestUser struct {
	TestEmail string
	TestId string
	TestIsAdmin bool
}

func (tu *TestUser) Email()string {
	return tu.TestEmail
}

func (tu *TestUser) Id()string {
	return tu.TestId
}

func (tu *TestUser) IsAdmin()bool {
	return tu.TestIsAdmin
}

func TestEmailHandler(t *testing.T){
	f,err := comp.CMSHandlerFactory(context.Background())

	if err != nil {
		t.Error(err)
	}

	resp := comp.ActionResponse{}

	match,err := f(&comp.EmailRequest{
		Email: comp.Email{
			Id: "test",
			Subject: "[CMS] Production | Environments Not Updating",
			Body: `[AWS] RE/MAX, LLC - REMAX CRM, (id: 3355) (update queued)
[AWS] RE/MAX, LLC - Motto Mortgage, (id: 3356) (update queued)
[AWS] RE/MAX, LLC - MAXCNTR, (id: 3357) (update queued)
[AWS] RE/MAX, LLC - Information Management, (id: 3359) (update queued)`,
			From: comp.EmailAddress{
				Address: "no-reply@cloud.datapipe.com",
				Name: "Datapipe Cloud Services",
			}},
		User: &TestUser{
			TestEmail: "test@test.com",
			TestId: "id",
			TestIsAdmin: false,
		},
		New: true,
	}, &resp)

	if err != nil {
		t.Error("Error writing object", err)
	}

	if match != true {
		t.Error("Not matched")
	}
}