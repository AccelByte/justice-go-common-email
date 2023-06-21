/*
 * Copyright (c) 2023 AccelByte Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and limitations under the License.
 *
 */

package emailsender

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/AccelByte/justice-go-common-email/object"
	"github.com/AccelByte/justice-go-common-email/platform"
	"github.com/AccelByte/justice-go-common-email/platform/mandrill"
	"github.com/AccelByte/justice-go-common-email/platform/sendgrid"
)

type StaticEmailSender struct {
	SenderPlatform platform.SenderPlatform
	FromAddress    string
	FromName       string
}

func NewStaticEmailSender() (*StaticEmailSender, error) {
	var senderPlatform, fromAddress, fromName string
	if senderPlatform = os.Getenv("APP_EMAIL_SENDER_NAME"); senderPlatform == "" {
		return nil, errors.New("APP_EMAIL_SENDER_NAME environment variable is not set")
	}
	if fromAddress = os.Getenv("FROM_EMAIL_ADDRESS"); fromAddress == "" {
		return nil, errors.New("FROM_EMAIL_ADDRESS environment variable is not set")
	}
	fromName = os.Getenv("FROM_EMAIL_NAME")

	emailSender := &StaticEmailSender{
		FromAddress: fromAddress,
		FromName:    fromName,
	}

	switch senderPlatform {
	case sendgrid.PlatformID:
		var apiKey, emailCategories string
		if apiKey = os.Getenv("SENDGRID_API_KEY"); apiKey == "" {
			return nil, errors.New("SENDGRID_API_KEY environment variable is not set")
		}
		emailCategories = os.Getenv("SENDGRID_EMAIL_CATEGORIES")

		emailSender.SenderPlatform = sendgrid.NewSendGridClient(apiKey, emailCategories)
	case mandrill.PlatformID:
		apiURL := "https://mandrillapp.com"
		smtpHost := "smtp.mandrillapp.com"
		smtpPort := 587
		var apiKey, smtpUsername, smtpPassword string

		if str := os.Getenv("MANDRILL_API_URL"); str != "" {
			apiURL = str
		}
		if str := os.Getenv("MANDRILL_SMTP_HOST"); str != "" {
			smtpHost = str
		}
		if s := os.Getenv("MANDRILL_SMTP_PORT"); s != "" {
			var err error
			smtpPort, err = strconv.Atoi(s)
			if err != nil {
				return nil, errors.New("MANDRILL_SMTP_PORT value must be an integer")
			}
		}

		apiKey = os.Getenv("MANDRILL_API_KEY")
		smtpUsername = os.Getenv("MANDRILL_USERNAME")
		smtpPassword = os.Getenv("MANDRILL_PASSWORD")

		if apiURL != "" && apiKey != "" { //nolint: gocritic
			emailSender.SenderPlatform = mandrill.NewMandrillClientWithAPIKey(apiURL, apiKey)
		} else if smtpHost != "" || smtpPort != 0 || smtpUsername != "" || smtpPassword != "" {
			emailSender.SenderPlatform = mandrill.NewMandrillClientWithSMTP(smtpHost, smtpPort, smtpUsername, smtpPassword)
		} else {
			return nil, errors.New("required mandrill environment variables is not set. For API Key: MANDRILL_API_URL, MANDRILL_API_KEY. For SMTP: MANDRILL_SMTP_HOST, MANDRILL_SMTP_PORT, MANDRILL_USERNAME, MANDRILL_PASSWORD")
		}
	default:
		return nil, fmt.Errorf("%s email sender platform is not valid", senderPlatform)
	}

	return emailSender, nil
}

func (e *StaticEmailSender) SendEmail(ctx context.Context, emailData object.EmailData) error {
	emailData.SetTemplateAdditionalData()
	emailData.From = e.FromAddress
	emailData.FromName = e.FromName
	return e.SenderPlatform.Send(ctx, emailData)
}
