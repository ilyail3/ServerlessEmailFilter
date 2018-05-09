package comp

import (
	"strings"
	"context"
	"regexp"
	"google.golang.org/appengine/datastore"
	"fmt"
	"time"
	"strconv"
)

type EmailRequest struct {
	Email Email
	User  AuthUser
	New   bool
}

func HandleEmail(handlers ...func(*EmailRequest, *ActionResponse) (bool, error)) func(*EmailRequest) (ActionResponse, error) {
	return func(request *EmailRequest) (ActionResponse, error) {
		response := ActionResponse{
			Type:   "action",
			Action: "do_nothing",
			EMail:  request.User.Email(),
			ID:     request.User.Id(),
			Admin:  request.User.IsAdmin(),
			New:    request.New,
		}

		for _, handler := range handlers {
			matched, err := handler(request, &response)

			if err != nil {
				return response, err
			}

			if matched {
				return response, nil
			}
		}

		return response, nil
	}
}

func AwsPendingConsolidationsHandler(request *EmailRequest, response *ActionResponse) (bool, error) {
	if request.Email.From.Address == "no-reply@cloud.datapipe.com" &&
		strings.HasPrefix(request.Email.Subject, "[AWS] Production | Pending Consolidations:") {

		response.Action = "delete"
		return true, nil
	}

	return false, nil
}

func AwsSubscriptionNotificationHandler(request *EmailRequest, response *ActionResponse) (bool, error) {
	if request.Email.From.Address == "no-reply-aws@amazon.com" &&
		request.Email.Subject == "Important Notification Regarding Your AWS Marketplace Subscription" {

		response.Action = "delete"
		return true, nil
	}

	return false, nil
}

func AzureDailyReportHandler(request *EmailRequest, response *ActionResponse) (bool, error) {
	if request.Email.From.Address == "no-reply@cms.dpcloud.com" &&
		strings.HasPrefix(request.Email.Subject, "CMS | Azure Daily Usage Audit Report:") {

		response.Action = "delete"
		return true, nil
	}

	return false, nil
}

func CAWelcomeHandler(request *EmailRequest, response *ActionResponse) (bool, error) {
	if request.Email.From.Address == "no-reply@cloud.datapipe.com" &&
		request.Email.Subject == "Welcome to Datapipe Cloud Analytics" {

		response.Action = "delete"
		return true, nil
	}

	return false, nil
}

type EnvironmentNotUpdated struct {
	Time   time.Time
	PaId   int
	UserId string
}

func CMSHandlerFactory(ctx context.Context) (func(*EmailRequest, *ActionResponse) (bool, error), error) {

	paIdRegex, err := regexp.Compile("\\(id: ([0-9]+)\\)")

	if err != nil {
		return nil, err
	}

	return func(r *EmailRequest, response *ActionResponse) (bool, error) {
		if (
			r.Email.From.Address == "no-reply@cloud.datapipe.com" ||
				r.Email.From.Address == "no-reply@cms.dpcloud.com") &&
			strings.HasPrefix(r.Email.Subject, "[CMS]") &&
			strings.Contains(r.Email.Subject, " | ") {

			withoutPrefix := r.Email.Subject[6:len(r.Email.Subject)]
			parts := strings.Split(withoutPrefix, " | ")

			// fmt.Printf("parts:%#v\n", parts)
			if parts[1] == "Environments Not Updating" {
				ids := paIdRegex.FindAllStringSubmatch(r.Email.Body, -1)

				for _, pair := range ids {
					key := datastore.NewKey(ctx, "enu", fmt.Sprintf(
						"%s-%s",
						r.User.Id(),
						pair[1]), 0, nil)

					paId, err := strconv.Atoi(pair[1])

					if err != nil {
						return false, err
					}

					env := EnvironmentNotUpdated{
						Time:   r.Email.ReceiveTime,
						PaId:   paId,
						UserId: r.User.Id(),
					}

					_, err = datastore.Put(ctx, key, &env)

					if err != nil {
						return false, err
					}
				}
			}

			return true, nil
		}

		return false, nil
	}, nil
}
