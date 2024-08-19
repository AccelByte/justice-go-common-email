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

package object

import (
	"time"

	"github.com/AccelByte/justice-go-common-email/constant"
)

type EmailData struct {
	/*
		Namespace that sending the email.
		In "configservice" mode: this namespace is used to fetch the email sender configuration from Config Service.
	*/
	Namespace string
	/*
		From Email Address.
		In "configservice" mode: from-address already configured in Config Service,
		but this field could be used to replace it if necessary.
	*/
	From string
	/*
		From Name.
		In "configservice" mode: from-name already configured in Config Service,
		but this field could be used to replace it if necessary.
	*/
	FromName     string
	To           string
	Subject      string
	ReplyTo      string
	XMCTemplate  string
	XMCMergeVars map[string]interface{}
	Categories   []string
	// CarbonCopy current this only supported for sendgrid and would need further update for mandrill
	CarbonCopy []string
}

func (d *EmailData) SetTemplateAdditionalData() {
	if d.XMCMergeVars == nil {
		d.XMCMergeVars = map[string]interface{}{}
	}
	// set default copyright year
	if _, exists := d.XMCMergeVars[constant.CopyrightYearTemplateKey]; !exists {
		d.XMCMergeVars[constant.CopyrightYearTemplateKey] = time.Now().Year()
	}
	// set language tag key
	if language := d.XMCMergeVars[constant.LanguageTagTemplateKey]; language != nil && language.(string) != "" {
		d.XMCMergeVars[language.(string)] = true
	}
}
