# justice-go-common-email

Go SDK for email sender functionality in AccelByte services.

## Usage

### Install

```
go get -u github.com/AccelByte/justice-go-common-email
```

### Importing

```go
"github.com/AccelByte/justice-go-common-email"
```
To create a new email sender client, use this function:
```go
// If using static configuration
emailSender, errEmailSender := emailsender.NewEmailSender(emailsender.StaticSource)

// If using config service configuration
emailSender, errEmailSender := emailsender.NewEmailSender(emailsender.ConfigServiceSource)
```

Example initialize email sender client by environment variable:
```go
var emailConfigSource emailsender.EmailConfigSource
if str := os.Getenv("APP_EMAIL_CONFIG_SOURCE"); str != "" { // using environment variable to decide the source config
	if str == string(emailsender.StaticSource) {
		emailConfigSource = emailsender.StaticSource
	} else if str == string(emailsender.ConfigServiceSource) {
		emailConfigSource = emailsender.ConfigServiceSource
	}
}
if emailConfigSource == nil {
	return errors.New("Source config is not valid")
}
emailSender, errEmailSender := emailsender.NewEmailSender(emailConfigSource)
if errEmailSender != nil {
	return errEmailSender
}
```

## Supported Email Sender Configuration
### Static Configuration

Read email sender configuration from environment variables.

Example initialization:
```go
emailSender, errEmailSender := emailsender.NewEmailSender(emailsender.StaticSource)
```

#### Environment Variables

| Environment Variable  | Description                                             |
|-----------------------|---------------------------------------------------------|
| APP_EMAIL_SENDER_NAME | Email sender platform. options: `sendgrid`, `mandrill`. |
| FROM_EMAIL_ADDRESS    | From email address, required.                           |
| FROM_EMAIL_NAME       | From email name.                                        |

##### If using `sendgrid` platform:</b>

| Environment Variable      | Description                 |
|---------------------------|-----------------------------|
| SENDGRID_API_KEY          | Sendgrid API Key, required. |
| SENDGRID_EMAIL_CATEGORIES | Sendgrid email categories.  |

##### If using `mandrill` platform:

There are 2 mode available when using `mandrill` platform: API and SMTP.
You could decide which mode you want to activate by configuring the environment variables below, then it will automatically use the configured one:

| Environment Variable      | Description                                          | Mode   |
|---------------------------|------------------------------------------------------|--------|
| MANDRILL_API_URL          | Mandrill API URL (default: https://mandrillapp.com). | API    |
| MANDRILL_API_KEY          | Mandrill API Key.                                    | API    |
| MANDRILL_SMTP_HOST        | Mandrill SMTP Host (default: smtp.mandrillapp.com).  | SMTP   |
| MANDRILL_SMTP_PORT        | Mandrill SMTP Port (default: 587).                   | SMTP   |
| MANDRILL_USERNAME         | Mandrill username.                                   | SMTP   |
| MANDRILL_PASSWORD         | Mandrill password.                                   | SMTP   |

### Config Service Configuration

Read email sender configuration from AccelByte Config Service.

Example initialization:
```go
emailSender, errEmailSender := emailsender.NewEmailSender(emailsender.ConfigServiceSource)
```

#### Environment Variables

| Environment Variable            | Description                                                                                                 |
|---------------------------------|-------------------------------------------------------------------------------------------------------------|
| APP_CONFIG_SERVICE_REMOTE_HOST  | Config Service host to fetch the email sender configuration (default: http://justice-config-service/config) |
| APP_CONFIG_SERVICE_CACHE_EXPIRE | Config Service cache expire in second (default: 60)                                                         |
| APP_EMAIL_SENDER_CACHE_EXPIRE   | Email sender platform cache expire in second (default: 60)                                                  |


## License

Copyright © 2023, AccelByte Inc. Released under the Apache License, Version 2.0