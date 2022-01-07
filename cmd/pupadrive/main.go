package main

import (
	_ "github.com/joho/godotenv/autoload"
	tb "gopkg.in/tucnak/telebot.v3"
	"pupadrive.go/internal/configs"
	"pupadrive.go/internal/handler"
	"pupadrive.go/internal/logger"
)

func main() {
	Bot, err := tb.NewBot(
		configs.GetTgConfig(),
	)

	if err != nil {
		logger.ErrorLogger.Panicln(err)
	}

	Bot.Handle("/start", handler.Start)

	Bot.Handle("/add", handler.AddMagnet)

	Bot.Handle("/yt", handler.DownloadVideo)

	logger.InfoLogger.Println("Starting bot...")
	Bot.Start()

}
