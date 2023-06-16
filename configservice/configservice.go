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

package configservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/AccelByte/justice-go-common-email/constant"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

const (
	getEmailSenderConfigurationPath = "%s/v1/admin/namespaces/%s/emailsender?includeEmailTemplates=true"
)

type APIProxy struct {
	Host  string
	Cache *cache.Cache
}

func NewConfigServiceProxy(host string, cacheExpireInSeconds int) (*APIProxy, error) {
	cacheExpireDuration := time.Duration(cacheExpireInSeconds)
	return &APIProxy{
		Host:  host,
		Cache: cache.New(cacheExpireDuration*time.Second, cacheExpireDuration*2*time.Second),
	}, nil
}

func (e *APIProxy) GetEmailSenderConfiguration(ctx context.Context, namespace string) (emailSender *EmailSenderConfiguration, err error) {
	result, found := e.Cache.Get(namespace)
	if found {
		emailSender = result.(*EmailSenderConfiguration)
	} else {
		emailSender, err = e.fetchEmailSenderConfiguration(ctx, namespace)
		if err != nil {
			if err == constant.ErrNotFound {
				return nil, nil
			}
			return nil, err
		}
		e.Cache.Set(namespace, emailSender, 0)
	}
	return emailSender, nil
}

func (e *APIProxy) fetchEmailSenderConfiguration(ctx context.Context, namespace string) (*EmailSenderConfiguration, error) {
	subCtx, cancel := context.WithTimeout(ctx, time.Second*constant.DefaultHTTPTimeoutInSeconds)
	defer cancel()

	accessToken, err := getAccessTokenFromContext(ctx)
	if err != nil {
		logrus.Error("Error get email sender config from config service: ", err)
		return nil, err
	}
	req, err := http.NewRequestWithContext(subCtx, http.MethodGet, fmt.Sprintf(getEmailSenderConfigurationPath, e.Host, namespace), nil)
	if err != nil {
		logrus.Error("Error get email sender config from config service: ", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	httpClient := &http.Client{
		Timeout: time.Second * constant.DefaultHTTPTimeoutInSeconds,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logrus.Error("Error get email sender config from config service: ", err)
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			errorEntity := &ErrorEntity{}
			err = json.Unmarshal(bodyBytes, errorEntity)
			if err != nil {
				logrus.Error("Error get email sender config from config service", err)
				return nil, errors.New(string(bodyBytes))
			}
			if errorEntity.ErrorCode == 20008 {
				logrus.Errorf("Email sender config for namespace %s is not found", namespace)
				return nil, constant.ErrNotFound
			}
		}
		logrus.Errorf("Error get email sender config from config service: %s", string(bodyBytes))
		return nil, errors.New(string(bodyBytes))
	}

	emailSenderConfig := &EmailSenderConfiguration{}
	err = json.Unmarshal(bodyBytes, emailSenderConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal json EmailSenderConfiguration: %v", err)
	}
	return emailSenderConfig, nil
}

func getAccessTokenFromContext(ctx context.Context) (string, error) {
	var accessToken string
	if token := ctx.Value(constant.ServiceAccessToken); token != nil {
		accessToken = token.(string)
	}
	if accessToken == "" {
		return "", errors.New("no access token specified inside context")
	}
	return accessToken, nil
}
