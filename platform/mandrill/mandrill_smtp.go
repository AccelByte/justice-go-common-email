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
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/smtp"

	"github.com/AccelByte/justice-go-common-email/object"
	"github.com/AccelByte/justice-go-common-email/platform"
	"github.com/sirupsen/logrus"
)

func NewMandrillClientWithSMTP(smtpHost string, smtpPort int, smtpUsername, smtpPassword string) platform.SenderPlatform {
	return &SMTPMailSender{
		Host:     smtpHost,
		Port:     smtpPort,
		Username: smtpUsername,
		Password: smtpPassword,
	}
}

func (e SMTPMailSender) Send(ctx context.Context, emailData object.EmailData) error {
	mergeVars, err := json.Marshal(emailData.XMCMergeVars)
	if err != nil {
		return err
	}

	from := mail.Address{Address: emailData.From, Name: emailData.FromName}
	toAddress := mail.Address{Address: emailData.To}

	headerKeys := []string{"Return-Path", "From", "To", "MIME-Version", "Content-Type", "Content-Transfer-Encoding"}
	header := make(map[string]string)
	header["Return-Path"] = from.Address
	header["From"] = from.String()
	header["To"] = toAddress.String()
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	if emailData.XMCTemplate != "" {
		header["X-MC-Template"] = emailData.XMCTemplate
		header["X-MC-MergeVars"] = string(mergeVars)
		headerKeys = append(headerKeys, "X-MC-Template", "X-MC-MergeVars")
	}
	if emailData.ReplyTo != "" {
		replyTo := mail.Address{Address: emailData.ReplyTo}
		header["Reply-To"] = replyTo.String()
		headerKeys = append(headerKeys, "Reply-To")
	}
	var msg string
	for _, key := range headerKeys {
		msg += fmt.Sprintf("%s: %s\r\n", key, header[key])
	}

	auth := smtp.PlainAuth("", e.Username, e.Password, e.Host)
	err = sendSMTPMail(
		fmt.Sprintf("%s:%d", e.Host, e.Port),
		auth,
		from.Address,
		[]string{toAddress.Address},
		[]byte(msg),
	)
	if err != nil {
		logrus.Errorf("Error send email to %s using Mandrill SMTP: %s", emailData.To, err)
	}
	return err
}

func sendSMTPMail(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	err := smtp.SendMail(addr, auth, from, to, msg)
	return err
}
