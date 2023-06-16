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

type ErrorEntity struct {
	ErrorCode    int    `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

type EmailTemplate struct {
	TemplateName string `json:"TemplateName"`
	TemplateID   string `json:"TemplateID"`
}

type EmailSenderConfiguration struct {
	Namespace             string           `json:"namespace"`
	FromAddress           string           `json:"fromAddress"`
	FromName              string           `json:"fromName"`
	APIKey                string           `json:"apiKey"`
	IsDomainAuthenticated bool             `json:"isDomainAuthenticated"`
	EmailTemplates        []*EmailTemplate `json:"emailTemplates,omitempty"`
}

func (d EmailSenderConfiguration) GetEmailTemplate(name string) *EmailTemplate {
	for _, t := range d.EmailTemplates {
		if t.TemplateName == name {
			return t
		}
	}
	return nil
}
