package telesend

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"regexp"
)

type TeleMessage interface {
	Send(bot *tgbotapi.BotAPI) (int, error)
}

type DeleteMessage struct {
	ChatId    int64
	MessageId int
}

func (t *DeleteMessage) Send(bot *tgbotapi.BotAPI) (int, error) {
	msg := tgbotapi.NewDeleteMessage(t.ChatId, t.MessageId)
	res, err := bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return res.MessageID, err
}

type ImageMessage struct {
	ChatId              int64
	FileBytes           []byte
	Caption             string
	DisableNotification bool
}

func (t *ImageMessage) Send(bot *tgbotapi.BotAPI) (int, error) {
	photoFileBytes := tgbotapi.FileBytes{
		Name:  "picture",
		Bytes: t.FileBytes,
	}
	msg := tgbotapi.NewPhotoUpload(t.ChatId, photoFileBytes)
	msg.DisableNotification = t.DisableNotification
	msg.Caption = t.Caption
	res, err := bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return res.MessageID, err
}

type TextMessage struct {
	ChatId              int64
	Text                string
	DisableNotification bool
}

func (t *TextMessage) Send(bot *tgbotapi.BotAPI) (int, error) {
	msg := tgbotapi.NewMessage(t.ChatId, removeUnsupportedTags(t.Text))
	msg.ParseMode = "html"
	msg.DisableNotification = t.DisableNotification
	res, err := bot.Send(msg)
	if err != nil {
		return 0, err
	}
	return res.MessageID, err
}

var replaceTags = regexp.MustCompile(`<[^>]*>`)

func removeAllTags(text string) string {
	return replaceTags.ReplaceAllString(text, "")
}

func removeUnsupportedTags(text string) string {
	for _, r := range replaces {
		text = r.from.ReplaceAllString(text, r.to)
	}
	return text
}

type regexpStruct struct {
	from *regexp.Regexp
	to   string
}

var replaces = []regexpStruct{
	{regexp.MustCompile(`<p[^>]*>`), ""},
	{regexp.MustCompile(`</p>`), "\n"},
	{regexp.MustCompile(`<span[^>]*>`), ""},
	{regexp.MustCompile(`</span>`), ""},
	{regexp.MustCompile(`<div[^>]*>`), ""},
	{regexp.MustCompile(`</div>`), ""},
	{regexp.MustCompile(`<h[1-9]>`), "<b>"},
	{regexp.MustCompile(`</h[1-9]>`), "</b>"},
	{regexp.MustCompile(`<li [^>]*>`), "* "},
	{regexp.MustCompile(`</li>`), ""},
	{regexp.MustCompile(`<ul [^>]*>`), "\n"},
	{regexp.MustCompile(`</ul>`), ""},
	{regexp.MustCompile(`</br>`), "\n"},
	{regexp.MustCompile(`<br>`), "\n"},
}
