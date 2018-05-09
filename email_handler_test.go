package main

import (
	"testing"
	"github.com/ilyail3/ServerlessEmailFilter/comp"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	"fmt"
	"time"
)

type TestUser struct {
	TestEmail   string
	TestId      string
	TestIsAdmin bool
}

func (tu *TestUser) Email() string {
	return tu.TestEmail
}

func (tu *TestUser) Id() string {
	return tu.TestId
}

func (tu *TestUser) IsAdmin() bool {
	return tu.TestIsAdmin
}

func TestEmailHandler(t *testing.T) {
	c, killContext, err := aetest.NewContext()

	if err != nil {
		t.Error(err)
		return
	}

	defer killContext()

	f, err := comp.CMSHandlerFactory(c)

	if err != nil {
		t.Error(err)
	}

	resp := comp.ActionResponse{}

	receiveTime := time.Now()

	match, err := f(&comp.EmailRequest{
		Email: comp.Email{
			Id:      "test",
			Subject: "[CMS] Production | Environments Not Updating",
			Body: `[AWS] RE/MAX, LLC - REMAX CRM, (id: 3355) (update queued)
[AWS] RE/MAX, LLC - Motto Mortgage, (id: 3356) (update queued)
[AWS] RE/MAX, LLC - MAXCNTR, (id: 3357) (update queued)
[AWS] RE/MAX, LLC - Information Management, (id: 3359) (update queued)`,
			From: comp.EmailAddress{
				Address: "no-reply@cloud.datapipe.com",
				Name:    "Datapipe Cloud Services",
			},
			ReceiveTime: receiveTime},
		User: &TestUser{
			TestEmail:   "test@test.com",
			TestId:      "id",
			TestIsAdmin: false,
		},
		New: true,
	}, &resp)

	if err != nil {
		t.Error("Error writing object", err)
		return
	}

	if match != true {
		t.Error("Not matched")
		return
	}

	for _, paId := range []int{3355, 3356, 3357, 3359} {
		enu := comp.EnvironmentNotUpdated{}
		err := datastore.Get(c, datastore.NewKey(c, "enu", fmt.Sprintf("id-%d", paId), 0, nil), &enu)

		// Return false, but don't shout about it
		if err == datastore.ErrNoSuchEntity {
			t.Errorf("missing record for pdId:%d", paId)
			return
		}

		if err != nil {
			t.Error(err)
			return
		}

		// Datastore rounds the number, which is expected, so compare in a reasonable resolution
		if enu.Time.Unix() != receiveTime.Unix() {
			t.Errorf("email time %v is different from record time %v", receiveTime, enu.Time)
			return
		}

		if enu.UserId != "id" {
			t.Errorf("bad user id %s", enu.UserId)
			return
		}

		if enu.PaId != paId {
			t.Errorf("bad paId %d", enu.PaId)
		}
	}
}
