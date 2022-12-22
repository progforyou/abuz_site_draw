package telesend

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type TeleBot struct {
	bot  *tgbotapi.BotAPI
	send chan TeleMessage
	exit chan bool
}

func NewTeleBot(token string) (*TeleBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot.Debug = true
	tBot := &TeleBot{
		bot:  bot,
		send: make(chan TeleMessage),
		exit: make(chan bool, 1),
	}
	go tBot.update()
	return tBot, nil
}

func (t *TeleBot) Send(to int64, message string) (int, error) {
	msg := tgbotapi.NewMessage(to, message)
	if res, err := t.bot.Send(msg); err != nil {
		return 0, err
	} else {
		return res.MessageID, nil
	}
}

func (t *TeleBot) SendMessage(message TeleMessage) (int, error) {
	return message.Send(t.bot)
}

func (t *TeleBot) Stop() {
	t.exit <- true
}

func (t *TeleBot) update() {
	log.Info().Msg("start tele-bot")
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 3
	updates := t.bot.GetUpdatesChan(updateConfig)
	for {
		select {
		case forSend := <-t.send:
			_, err := forSend.Send(t.bot)
			if err != nil {
				log.Error().Err(err).Msg("send-error tele-bot")
			}
		case update := <-updates:
			if update.Message != nil {
				log.Info().Str("type", "Message").Int64("chat_id", update.Message.Chat.ID).Msg(update.Message.Text)
			}
			if update.ChannelPost != nil {
				log.Info().Str("type", "ChannelPost").Int64("chat_id", update.ChannelPost.Chat.ID).Msg(update.ChannelPost.Text)
			}
		case <-t.exit:
			log.Info().Msg("stop tele-bot")
			return
		}
	}
}
