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
	"fmt"

	"github.com/AccelByte/justice-go-common-email/object"
)

type EmailConfigSource string

const (
	StaticSource        EmailConfigSource = "static"
	ConfigServiceSource EmailConfigSource = "configservice"
)

type EmailSender interface {
	SendEmail(ctx context.Context, emailData object.EmailData) error
}

func NewEmailSender(configSource EmailConfigSource) (EmailSender, error) {
	switch configSource {
	case StaticSource:
		return NewStaticEmailSender()
	case ConfigServiceSource:
		return NewConfigServiceEmailSender()
	default:
		return nil, fmt.Errorf("unsupported %s config source", configSource)
	}
}
