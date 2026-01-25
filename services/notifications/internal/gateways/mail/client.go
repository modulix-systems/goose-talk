package mailclient

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"path"
	"strings"
	"time"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"
	"github.com/modulix-systems/goose-talk/internal/utils"
)

type SmtpMailClient struct {
	addr    string
	from    string
	auth    smtp.Auth
	appName string
	appUrl  string
}

type TemplateData[T any] struct {
	Payload T
	AppName string
	AppUrl  string
	Year    int
}

func New(host, port, username, password, appName, appUrl string) *SmtpMailClient {
	addr := net.JoinHostPort(host, port)
	auth := smtp.PlainAuth("", username, password, host)

	return &SmtpMailClient{from: username, addr: addr, auth: auth, appName: appName, appUrl: appUrl}
}

func send[T any](client *SmtpMailClient, payload T, to, templateName, subject string) error {
	var messageBody bytes.Buffer

	templateData := TemplateData[T]{
		payload,
		client.appName,
		client.appUrl,
		time.Now().Year(),
	}
	tmpl, err := template.ParseFiles(path.Join(utils.FindRootPath(), "internal", "gateways", "mail", "templates", templateName))
	if err != nil {
		return fmt.Errorf("mailclient - send - template.ParseFile: %w", err)
	}
	if err = tmpl.Execute(&messageBody, templateData); err != nil {
		return fmt.Errorf("mailclient - send - tmpl.Execute: %w", err)
	}

	messageParts := []string{
		fmt.Sprintf("From: %s", client.from),
		fmt.Sprintf("Subject: %s", subject),
		fmt.Sprintf("To: %s", to),
		"Content-Type: text/html",
		"",
		messageBody.String(),
	}
	message := strings.Join(messageParts, "\r\n")

	if err := smtp.SendMail(client.addr, client.auth, client.from, []string{to}, []byte(message)); err != nil {
		return fmt.Errorf("mailclient - send - smtp.SendMail: %w", err)
	}

	return nil
}

func (c *SmtpMailClient) SendSignUpNotice(ctx context.Context, to string, data notifications.SignUpNotice, lang notifications.Language) error {
	return send(c, data, to, "signup.html", getEmailSubject(notifications.EMAIL_TYPE_SIGN_UP, lang))
}
func (c *SmtpMailClient) SendLoginNewDeviceNotice(ctx context.Context, to string, data notifications.LoginNewDeviceNotice, lang notifications.Language) error {
	return send(c, data, to, "login_new_device.html", getEmailSubject(notifications.EMAIL_TYPE_LOGIN_NEW_DEVICE, lang))
}
func (c *SmtpMailClient) SendAccountDeactivatedNotice(ctx context.Context, to string, data notifications.AccountDeactivatedNotice, lang notifications.Language) error {
	return send(c, data, to, "account_deactivated.html", getEmailSubject(notifications.EMAIL_TYPE_ACCOUNT_DEACTIVATED, lang))
}
func (c *SmtpMailClient) SendVerifyEmailNotice(ctx context.Context, to string, data notifications.EmailVerifyNotice, lang notifications.Language) error {
	return send(c, data, to, "verify_email.html", getEmailSubject(notifications.EMAIL_TYPE_VERIFY_EMAIL, lang))
}
func (c *SmtpMailClient) SendConfirmEmailTwoFaNotice(ctx context.Context, to string, data notifications.EmailTwoFaNotice, lang notifications.Language) error {
	return send(c, data, to, "email_two_fa.html", getEmailSubject(notifications.EMAIL_TYPE_EMAIL_TWO_FA, lang))
}
func (c *SmtpMailClient) SendConfirmedTwoFaNotice(ctx context.Context, to string, data notifications.TwoFaConfirmedNotice, lang notifications.Language) error {
	return send(c, data, to, "two_fa_confirmed.html", getEmailSubject(notifications.EMAIL_TYPE_TWO_FA_CONFIRMED, lang))
}
