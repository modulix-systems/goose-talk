package notifications

import "encoding/json"

type EmailType string

var (
	EMAIL_TYPE_SIGN_UP             EmailType = "signup"
	EMAIL_TYPE_VERIFY_EMAIL        EmailType = "verify_email"
	EMAIL_TYPE_ACCOUNT_DEACTIVATED EmailType = "account_deactivated"
	EMAIL_TYPE_LOGIN_NEW_DEVICE    EmailType = "login_new_device"
	EMAIL_TYPE_EMAIL_TWO_FA        EmailType = "email_two_fa"
	EMAIL_TYPE_TWO_FA_CONFIRMED    EmailType = "two_fa_confirmed"
)

type Language string

var (
	LANGUAGE_EN Language = "EN"
	LANGUAGE_RU Language = "RU"
)

type EmailMessage struct {
	Type     EmailType
	Language Language
	To       string
	Data     json.RawMessage
}

type SignUpNotice struct {
	Username string
}

type EmailTwoFaNotice struct {
	Username string
	Code     string
}

type EmailVerifyNotice struct {
	Username string
	Code     string
}

type AccountDeactivatedNotice struct {
	Username string
}

type TwoFaConfirmedNotice struct {
	Username    string
	TwoFaMethod string
}

type LoginNewDeviceNotice struct {
	Username   string
	IpAddr     string
	DeviceInfo string
	Location   string
}
