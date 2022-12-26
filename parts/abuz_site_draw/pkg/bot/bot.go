package bot

import (
	"abuz_site_draw/parts/abuz_site_draw/pkg/data"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

func StartTelegramBot(telegramToken, ws string, c *data.Controllers) {
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

			if update.Message.ConnectedWebsite == ws {
				log.Info().Str("tg name", update.Message.Chat.UserName).Msg("login as")
			}

		}
	}()
}
