package handler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/dustin/go-humanize"
	tb "gopkg.in/tucnak/telebot.v3"
	"pupadrive.go/internal/configs"
	"pupadrive.go/internal/gdrive"
	"pupadrive.go/internal/logger"
)

var (
	torrentClient *torrent.Client
)

func init() {
	c, err := torrent.NewClient(configs.GetTorrentConfig())
	if err != nil {
		logger.ErrorLogger.Panicln(err)
	}
	torrentClient = c
}

func AddMagnet(c tb.Context) error {
	magnet := c.Message().Payload

	if !strings.HasPrefix(magnet, "magnet:?") {
		logger.ErrorLogger.Println(fmt.Sprintln(c.Message(), "Please enter magnet link"))
		_, err := c.Bot().Send(c.Chat(), "Please enter magnet link")
		return err
	}

	err := downloadMagnet(c, magnet)
	if err != nil {
		logger.ErrorLogger.Println(err)
		c.Bot().Send(c.Chat(), fmt.Sprintf("Error: %s", err.Error()))
		return err
	}
	return nil
}

func downloadMagnet(c tb.Context, magnet string) error {
	t, err := torrentClient.AddMagnet(magnet)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}

	m, _ := c.Bot().Send(c.Chat(), "Fetching metadata...")

	<-t.GotInfo()
	info := t.Info()

	exists, err := gdrive.DriveClient.CheckFolderExist(info.Name, "")
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}

	if exists != "" {
		r := &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{
			{
				tb.InlineButton{
					Unique: "view",
					Text:   "View",
					URL:    fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", exists),
				},
			},
		}}

		c.Bot().Edit(m, fmt.Sprintf("*%s*\n_Already uploaded_",
			info.Name,
		), &tb.SendOptions{
			ParseMode:   tb.ModeMarkdown,
			ReplyMarkup: r,
		},
		)
	}

	total := info.TotalLength()
	t.DownloadAll()
	started := time.Now()

	for {
		if t.Complete.Bool() {
			break
		}

		stats := t.Stats()
		current := stats.BytesReadUsefulData.Int64()
		c.Bot().Edit(m, fmt.Sprintf("*%s*\n_Downloading_ %.2f%%\n\n%s of %s done.\n\n⬇️ %s/s | P: %d | S: %d", info.Name,
			float32(current)/float32(total)*100,
			humanize.Bytes(uint64(current)),
			humanize.Bytes(uint64(total)),
			humanize.Bytes(uint64(float64(current)/time.Since(started).Seconds())),
			stats.ActivePeers,
			stats.ConnectedSeeders,
		), tb.ModeMarkdown)

		time.Sleep(time.Second)
	}

	t.Drop()

	uploadFinished := false
	uploadCurrent := int64(0)
	uploadStarted := time.Now()

	parentID, err := gdrive.DriveClient.CreateFolder(info.Name, "")
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}

	torrentDir := fmt.Sprintf("./downloads/%s", t.InfoHash().String())

	go func() {
		err = gdrive.DriveClient.UploadFolder(torrentDir, parentID, func(current, total int64) {
			uploadCurrent = current
		})
		uploadFinished = true
	}()

	for !uploadFinished {
		c.Bot().Edit(m, fmt.Sprintf("*%s*\n_Uploading_ (%.2f %%) %s of %s\n\n⬆️ %s/s",
			info.Name,
			float32(uploadCurrent)/float32(total)*100,
			humanize.Bytes(uint64(uploadCurrent)),
			humanize.Bytes(uint64(total)),
			humanize.Bytes(uint64(float64(uploadCurrent)/time.Since(uploadStarted).Seconds())),
		), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		})
		time.Sleep(time.Second)
	}

	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}

	os.RemoveAll(torrentDir)

	r := &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{
		{
			tb.InlineButton{
				Unique: "view",
				Text:   "View",
				URL:    fmt.Sprintf("https://drive.google.com/drive/u/0/folders/%s", parentID),
			},
		},
	}}

	c.Bot().Edit(m, fmt.Sprintf("*%s*\n_Finished_ (%s)",
		info.Name,
		humanize.Bytes(uint64(total)),
	), &tb.SendOptions{
		ParseMode:   tb.ModeMarkdown,
		ReplyMarkup: r,
	},
	)

	return nil
}
