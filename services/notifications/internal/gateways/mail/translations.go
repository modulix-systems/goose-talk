package mailclient

import "github.com/modulix-systems/goose-talk/contracts/rmqcontracts/notifications"

type TranslationTable = map[notifications.Language]string

func getEmailSubject(typ notifications.EmailType, lang notifications.Language) string {
	if lang == "" {
		lang = notifications.LANGUAGE_EN
	}

	translations := map[notifications.EmailType]TranslationTable{
		notifications.EMAIL_TYPE_VERIFY_EMAIL: {
			notifications.LANGUAGE_EN: "Confirm your email",
			notifications.LANGUAGE_RU: "Подтвердите электронную почту",
		},

		notifications.EMAIL_TYPE_SIGN_UP: {
			notifications.LANGUAGE_EN: "Welcome! Complete your sign-up",
			notifications.LANGUAGE_RU: "Добро пожаловать! Завершите регистрацию",
		},

		notifications.EMAIL_TYPE_ACCOUNT_DEACTIVATED: {
			notifications.LANGUAGE_EN: "Your account has been deactivated",
			notifications.LANGUAGE_RU: "Ваш аккаунт был деактивирован",
		},

		notifications.EMAIL_TYPE_LOGIN_NEW_DEVICE: {
			notifications.LANGUAGE_EN: "New login from a new device",
			notifications.LANGUAGE_RU: "Вход с нового устройства",
		},

		notifications.EMAIL_TYPE_EMAIL_TWO_FA: {
			notifications.LANGUAGE_EN: "Your two-factor authentication code",
			notifications.LANGUAGE_RU: "Код двухфакторной аутентификации",
		},

		notifications.EMAIL_TYPE_TWO_FA_CONFIRMED: {
			notifications.LANGUAGE_EN: "Two-factor authentication enabled",
			notifications.LANGUAGE_RU: "Двухфакторная аутентификация включена",
		},
	}

	return translations[typ][lang]
}
