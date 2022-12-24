package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

func StartTelegramBot(telegramToken string) {
	bot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatal().Err(err).Msg("fail add bot")
		return
	}
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	go func() {
		updates := bot.GetUpdatesChan(u)

		for update := range updates {
			if update.Message == nil { // ignore non-Message updates
				continue
			}

			if update.Message.Text == "" {
				continue
			}
			log.Info().Int64("chat ID", update.Message.Chat.ID).Msg("Message from")

		}
	}()
}
