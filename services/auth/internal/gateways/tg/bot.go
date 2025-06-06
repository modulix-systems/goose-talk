package tg

type TgAPI struct {
	baseUrl string
	botUrl  string
}

func New(botToken string) *TgAPI {
	return &TgAPI{
		baseUrl: "https://api.telegram.org/bot" + botToken,
		botUrl:  "https://t.me/goose_talk_2fa_bot",
	}
}

func (t *TgAPI) GetStartLinkWithCode(code string) string {
	return t.botUrl + "/?start=" + code
}
