package tg

type TgAPI struct {
	baseUrl string
}

func (t *TgAPI) GetStartLinkWithCode(code string) string {
	return "https://t.me/goose_talk_2fa_bot/?start=" + code
}
