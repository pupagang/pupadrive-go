package configs

import (
	"net/http"
	"net/url"
	"os"

	"pupadrive.go/internal/logger"
)

type configs struct {
	Proxy             string
	TeamDriveID       string
	BotToken          string
	YoutubeProxy      string
	TeamDriveFolderID string
}

var config *configs

func init() {
	config = &configs{
		Proxy:             os.Getenv("PROXY"),
		TeamDriveID:       os.Getenv("DRIVE_ID"),
		BotToken:          os.Getenv("BOT_TOKEN"),
		YoutubeProxy:      os.Getenv("HTTP_PROXY"),
		TeamDriveFolderID: os.Getenv("GOOGLE_DRIVE_FOLDER"),
	}
}

func GetProxy() func(*http.Request) (*url.URL, error) {
	proxyURL, err := url.Parse(config.Proxy)
	if err != nil {
		logger.ErrorLogger.Panic(err.Error())
	}
	return http.ProxyURL(proxyURL)
}

func GetTeamDriveID() string {
	return config.TeamDriveID
}

func GetBotToken() string {
	return config.BotToken
}

func GetYoutubeProxy() func(*http.Request) (*url.URL, error) {
	proxyURL, err := url.Parse(config.Proxy)
	if err != nil {
		logger.ErrorLogger.Panic(err.Error())
	}
	return http.ProxyURL(proxyURL)
}
func GetTeamDriveFolderID() string {
	return config.TeamDriveFolderID
}
