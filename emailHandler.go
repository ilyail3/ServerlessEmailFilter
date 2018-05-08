package main

import (
	"strings"
)

type EmailRequest struct{
	Email Email
	User AuthUser
	New bool
}

func HandleEmail(handlers ...func(*EmailRequest, *ActionResponse) bool) func(*EmailRequest) ActionResponse {
	return func(request *EmailRequest) ActionResponse {
		response := ActionResponse{
			Type:   "action",
			Action: "do_nothing",
			EMail:  request.User.Email(),
			ID:     request.User.Id(),
			Admin:  request.User.IsAdmin(),
			New:    request.New,
		}

		for _, handler := range handlers {
			if handler(request, &response) {
				return response
			}
		}

		return response
	}
}

func AwsPendingConsolidationsHandler(request *EmailRequest, response *ActionResponse) bool {
	if request.Email.From.Address == "no-reply@cloud.datapipe.com" &&
		strings.HasPrefix(request.Email.Subject, "[AWS] Production | Pending Consolidations:") {

		response.Action = "delete"
		return true
	}

	return false
}

func AwsSubscriptionNotificationHandler(request *EmailRequest, response *ActionResponse) bool {
	if request.Email.From.Address == "no-reply-aws@amazon.com" &&
		request.Email.Subject == "Important Notification Regarding Your AWS Marketplace Subscription" {

		response.Action = "delete"
		return true
	}

	return false
}

func AzureDailyReportHandler(request *EmailRequest, response *ActionResponse) bool {
	if request.Email.From.Address == "no-reply@cms.dpcloud.com" &&
		strings.HasPrefix(request.Email.Subject, "CMS | Azure Daily Usage Audit Report:") {

		response.Action = "delete"
		return true
	}

	return false
}

func CAWelcomeHandler(request *EmailRequest, response *ActionResponse) bool {
	if request.Email.From.Address == "no-reply@cloud.datapipe.com" &&
		request.Email.Subject == "Welcome to Datapipe Cloud Analytics" {

		response.Action = "delete"
		return true
	}

	return false
}