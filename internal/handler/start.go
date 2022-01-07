package handler

import (
	tb "gopkg.in/tucnak/telebot.v3"
)

func Start(c tb.Context) error {
	err := c.Send(
		`Welcome to PupaDrive!
		 
Usage: 
/add magnet link to add a torrent
/yt youtube link to download a video`)
	if err != nil {
		return err
	}
	return nil
}
