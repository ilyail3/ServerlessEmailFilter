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
	err,f := comp.CMSHandlerFactory(context.Background())

	if err != nil {
		t.Error(err)
	}

	resp := comp.ActionResponse{}

	match := f(&comp.EmailRequest{
		Email: comp.Email{
			Id: "test",
			Subject: "[CMS] Production | Datapipe One CI Discrepancy Report",
			Body: "",
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

	if match != true {
		t.Error("Not matched")
	}
}