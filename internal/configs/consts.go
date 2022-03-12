package configs

import (
	"os"
)

type configs struct {
	TeamDriveID       string
	BotToken          string
	TeamDriveFolderID string
}

var config *configs

func init() {
	config = &configs{
		TeamDriveID:       os.Getenv("DRIVE_ID"),
		BotToken:          os.Getenv("BOT_TOKEN"),
		TeamDriveFolderID: os.Getenv("GOOGLE_DRIVE_FOLDER"),
	}
}

func GetTeamDriveID() string {
	return config.TeamDriveID
}

func GetBotToken() string {
	return config.BotToken
}

func GetTeamDriveFolderID() string {
	return config.TeamDriveFolderID
}
