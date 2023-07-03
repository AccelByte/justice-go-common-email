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
	"time"

	"github.com/AccelByte/justice-go-common-email/configservice"
	"github.com/AccelByte/justice-go-common-email/object"
	"github.com/AccelByte/justice-go-common-email/platform"
	"github.com/AccelByte/justice-go-common-email/platform/sendgrid"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

type ConfigServiceEmailSender struct {
	ConfigServiceProxy  *configservice.APIProxy
	SenderPlatformCache *cache.Cache
}

func NewConfigServiceEmailSender() (*ConfigServiceEmailSender, error) {
	// init config service proxy
	host := "http://justice-config-service/config"
	configServiceCacheExpire := 60
	if str := os.Getenv("APP_CONFIG_SERVICE_REMOTE_HOST"); str != "" {
		host = str
	}
	if s := os.Getenv("APP_CONFIG_SERVICE_CACHE_EXPIRE"); s != "" {
		var err error
		configServiceCacheExpire, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("APP_CONFIG_SERVICE_CACHE_EXPIRE value must be an integer")
		}
	}
	configServiceProxy, err := configservice.NewConfigServiceProxy(host, configServiceCacheExpire)
	if err != nil {
		return nil, fmt.Errorf("fail initialize Config Service Proxy: %s", err.Error())
	}

	// init cache
	senderCacheExpire := 60
	if s := os.Getenv("APP_EMAIL_SENDER_CACHE_EXPIRE"); s != "" {
		var err error
		senderCacheExpire, err = strconv.Atoi(s)
		if err != nil {
			return nil, errors.New("APP_EMAIL_SENDER_CACHE_EXPIRE value must be an integer")
		}
	}
	senderPlatformCache := cache.New(time.Duration(senderCacheExpire)*time.Second, time.Duration(senderCacheExpire)*2*time.Second)

	return &ConfigServiceEmailSender{
		ConfigServiceProxy:  configServiceProxy,
		SenderPlatformCache: senderPlatformCache,
	}, nil
}

func (e *ConfigServiceEmailSender) SendEmail(ctx context.Context, emailData object.EmailData) error {
	if emailData.Namespace == "" {
		return errors.New("namespace is not specified yet")
	}

	emailSenderConfiguration, err := e.ConfigServiceProxy.GetEmailSenderConfiguration(ctx, emailData.Namespace)
	if err != nil {
		logrus.Errorf("fail get email sender configuration. error: %v", err)
		return err
	}
	if emailSenderConfiguration == nil {
		logrus.Errorf("email sender configuration for namespace %s is not found", emailData.Namespace)
		return ErrConfigurationNotFound
	}
	if !emailSenderConfiguration.IsDomainAuthenticated {
		logrus.Errorf("email sender domain for namespace %s is not authenticated yet", emailData.Namespace)
		return ErrConfigurationNotValid
	}

	if emailData.From == "" {
		emailData.From = emailSenderConfiguration.FromAddress
	}
	if emailData.FromName == "" {
		emailData.FromName = emailSenderConfiguration.FromName
	}
	if emailData.XMCTemplate != "" {
		// if XMCTemplate is specified, we will try to replace it with email template from Sender Configuration,
		// but if email template not found, we will keep use the specified value.

		emailTemplate := emailSenderConfiguration.GetEmailTemplate(emailData.XMCTemplate)
		if emailTemplate != nil {
			emailData.XMCTemplate = emailTemplate.TemplateID
		}
	}
	emailData.SetTemplateAdditionalData()

	senderPlatform := e.getSenderPlatform(emailSenderConfiguration.APIKey)
	if senderPlatform == nil {
		logrus.Errorf("sender platform for namespace %s is not exist", emailData.Namespace)
		return ErrSenderPlatformNotExist
	}
	return senderPlatform.Send(ctx, emailData)
}

func (e *ConfigServiceEmailSender) getSenderPlatform(apiKey string) (senderPlatform platform.SenderPlatform) {
	result, found := e.SenderPlatformCache.Get(apiKey)
	if found {
		senderPlatform = result.(platform.SenderPlatform)
	} else {
		// We only supports SendGrid for now
		senderPlatform = sendgrid.NewSendGridClient(apiKey, "")
		e.SenderPlatformCache.Set(apiKey, senderPlatform, 0)
	}
	return senderPlatform
}
