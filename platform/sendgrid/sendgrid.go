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

package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/AccelByte/justice-go-common-email/constant"
	"github.com/AccelByte/justice-go-common-email/object"
	"github.com/AccelByte/justice-go-common-email/platform"
	"github.com/sirupsen/logrus"
)

const (
	PlatformID = "sendgrid"

	apiHost       = "https://api.sendgrid.com"
	sendEmailPath = "/v3/mail/send"
)

type MailSender struct {
	Host                   string
	APIKey                 string
	DefaultEmailCategories string
}

type mail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type attachment struct {
	Content  string `json:"content"`
	Types    string `json:"type"`
	Filename string `json:"filename"`
}

type emailPayload struct {
	Subject          string            `json:"subject"`
	From             mail              `json:"from"`
	ReplyTo          *mail             `json:"reply_to,omitempty"`
	Personalizations []personalization `json:"personalizations"`
	TemplateID       string            `json:"template_id"`
	Attachments      []attachment      `json:"attachments"`
	Categories       []string          `json:"categories,omitempty"`
}

type personalization struct {
	To                  []mail                 `json:"to"`
	CC                  []mail                 `json:"cc"`
	DynamicTemplateData map[string]interface{} `json:"dynamic_template_data"`
}

func NewSendGridClient(apiKey, emailCategories string) platform.SenderPlatform {
	return &MailSender{
		Host:                   apiHost,
		APIKey:                 apiKey,
		DefaultEmailCategories: emailCategories,
	}
}

func (e MailSender) Send(ctx context.Context, emailData object.EmailData) error {
	// set default email categories
	var emailCategories []string
	if e.DefaultEmailCategories != "" {
		emailCategories = strings.Split(e.DefaultEmailCategories, ",")
	}
	if len(emailData.Categories) > 0 {
		emailCategories = append(emailCategories, emailData.Categories...)
	}

	personalizations := []personalization{
		{
			To:                  []mail{{Email: emailData.To}},
			DynamicTemplateData: emailData.XMCMergeVars,
		},
	}
	CCs := make([]mail, 0)
	for _, ccEmailData := range emailData.CarbonCopy {
		CCs = append(CCs, mail{Email: ccEmailData})
	}
	personalizations = append(personalizations, personalization{CC: CCs})
	payload := &emailPayload{
		Subject:          emailData.Subject,
		From:             mail{Email: emailData.From, Name: emailData.FromName},
		Personalizations: personalizations,
		TemplateID:       emailData.XMCTemplate,
		Categories:       emailCategories,
	}
	if emailData.ReplyTo != "" {
		payload.ReplyTo = &mail{Email: emailData.ReplyTo}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(payloadBytes)

	subCtx, cancel := context.WithTimeout(ctx, time.Second*constant.DefaultHTTPTimeoutInSeconds)
	defer cancel()
	req, err := http.NewRequestWithContext(subCtx, http.MethodPost, e.Host+sendEmailPath, body)
	if err != nil {
		logrus.Errorf("Error send email to %s using sendgrid: %s", emailData.To, err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: time.Second * constant.DefaultHTTPTimeoutInSeconds,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Errorf("Error send email to %s using sendgrid: %s", emailData.To, err)
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusAccepted {
		errorsResponseBody, errReadResp := ioutil.ReadAll(resp.Body)
		if errReadResp != nil {
			return errReadResp
		}
		logrus.Errorf("Error send email to %s using sendgrid: %s", emailData.To, string(errorsResponseBody))
		return errors.New(string(errorsResponseBody))
	}
	return nil
}
