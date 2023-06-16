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

package mandrill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/AccelByte/justice-go-common-email/constant"
	"github.com/AccelByte/justice-go-common-email/object"
	"github.com/AccelByte/justice-go-common-email/platform"
	"github.com/sirupsen/logrus"
)

const sendEmailPath = "/api/1.0/messages/send-template.json"

type mailTo struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type"`
}

type attachment struct {
	Types   string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

type mergeVar struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type message struct {
	Subject         string       `json:"subject"`
	FromEmail       string       `json:"from_email"`
	FromName        string       `json:"from_name"`
	To              []mailTo     `json:"to"`
	GlobalMergeVars []mergeVar   `json:"global_merge_vars"`
	Attachment      []attachment `json:"attachments"`
}

type emailPayload struct {
	Key             string  `json:"key"`
	TemplateName    string  `json:"template_name"`
	TemplateContent string  `json:"template_content"`
	Message         message `json:"message"`
	Async           bool    `json:"async"`
}

func NewMandrillClientWithAPIKey(apiURL, apiKey string) platform.SenderPlatform {
	return &MailSender{
		Host:   apiURL,
		APIKey: apiKey,
	}
}

func (e MailSender) Send(ctx context.Context, emailData object.EmailData) error {
	mergeVars := convertToMergeVars(emailData.XMCMergeVars)
	msg := message{
		Subject:   emailData.Subject,
		FromEmail: emailData.From,
		FromName:  emailData.FromName,
		To: []mailTo{
			{
				Email: emailData.To,
				Type:  "to",
			},
		},
		GlobalMergeVars: mergeVars,
	}
	payload := &emailPayload{
		Key:          e.APIKey,
		TemplateName: emailData.XMCTemplate,
		Message:      msg,
		Async:        false,
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
		logrus.Errorf("Error send email to %s using Mandrill API: %s", emailData.To, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: time.Second * constant.DefaultHTTPTimeoutInSeconds,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Errorf("Error send email to %s using Mandrill API: %s", emailData.To, err)
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		errorsResponseBody, errReadResp := ioutil.ReadAll(resp.Body)
		if errReadResp != nil {
			return errReadResp
		}
		logrus.Errorf("Error send email to %s using Mandrill API: %s", emailData.To, string(errorsResponseBody))
		return errors.New(string(errorsResponseBody))
	}
	return nil
}

func convertToMergeVars(xmcMergeVars map[string]interface{}) []mergeVar {
	mergeVars := make([]mergeVar, 0)
	if xmcMergeVars != nil { // nolint: gosimple
		for k, v := range xmcMergeVars {
			mergeVars = append(mergeVars, mergeVar{
				Name:    k,
				Content: fmt.Sprintf("%v", v),
			})
		}
	}
	return mergeVars
}
