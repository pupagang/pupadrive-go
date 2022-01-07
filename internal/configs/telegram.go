package configs

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v3"
)

func GetTgConfig() tb.Settings {
	return tb.Settings{
		Token:  GetBotToken(),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	}
}
