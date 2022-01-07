package handler

import (
	"net/http"
	"os"
	"os/exec"
	"strings"

	"fmt"
	"io"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/kkdai/youtube/v2"
	tb "gopkg.in/tucnak/telebot.v3"
	"pupadrive.go/internal/configs"
	"pupadrive.go/internal/gdrive"
	"pupadrive.go/internal/logger"
)

var (
	httpClient = &http.Client{}
	client     = &youtube.Client{
		HTTPClient: httpClient,
	}
)

type PassThru struct {
	io.ReadCloser
	current uint64
}

func (pt *PassThru) Read(p []byte) (int, error) {
	n, err := pt.ReadCloser.Read(p)
	if err == nil {
		pt.current += uint64(n)
	}

	return n, err
}

func (pt *PassThru) Close() error {
	return pt.ReadCloser.Close()
}

func DownloadVideo(c tb.Context) error {
	url := c.Message().Payload

	if !strings.Contains(url, "youtu") {
		logger.ErrorLogger.Println(fmt.Sprintln(c.Message(), "Please enter a valid youtube url"))
		_, err := c.Bot().Reply(c.Message(), "Please enter a valid youtube url")
		return err
	}

	message, _ := c.Bot().Send(c.Chat(), "Downloading...")

	video, err := client.GetVideo(url)
	if err != nil {
		logger.ErrorLogger.Println(err)
		_, err := c.Bot().Reply(message, fmt.Sprintf("Error: %s", err.Error()))
		return err

	}

	fileName := genFilename(video.Title)
	fileExists, err := gdrive.DriveClient.CheckFileExist(fileName)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}

	if fileExists != "" {
		r := &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{
			{
				tb.InlineButton{
					Unique: "download",
					Text:   "Download",
					URL:    fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", fileExists),
				},
			},
		}}

		c.Bot().Edit(message, fmt.Sprintf("*%s*\n_Already uploaded_",
			fileName,
		), &tb.SendOptions{
			ParseMode:   tb.ModeMarkdown,
			ReplyMarkup: r,
		},
		)
		return err
	}

	if len(video.Formats) == 0 {
		logger.ErrorLogger.Println(err)
		_, err := c.Bot().Reply(c.Message(), "No video formats found")
		return err
	}

	video.Formats.Sort()
	videoFormat := video.Formats.Type("video")[0]
	var driveID string
	var fileSize int64

	if videoFormat.AudioChannels == 0 {
		audioFormat := video.Formats.Type("audio")[0]
		driveID, fileSize, err = downloadVideoAudio(video, &videoFormat, &audioFormat, c, message, fileName)

	} else {
		driveID, fileSize, err = downloadVideoOnly(video, &videoFormat, c, message)
	}

	if err != nil {
		logger.ErrorLogger.Println(err)
		_, err := c.Bot().Reply(c.Message(), fmt.Sprintf("Error: %s", err.Error()))
		return err
	}

	r := &tb.ReplyMarkup{InlineKeyboard: [][]tb.InlineButton{
		{
			tb.InlineButton{
				Unique: "view",
				Text:   "View",
				URL:    fmt.Sprintf("https://drive.google.com/file/d/%s/view", driveID),
			},
			tb.InlineButton{
				Unique: "download",
				Text:   "Download",
				URL:    fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", driveID),
			},
		},
	}}

	c.Bot().Edit(message, fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>\n<i>Finished</i>\n\n🎞 %s | 🗄 %s",
		fmt.Sprintf("https://youtu.be/%s", video.ID),
		video.Title,
		videoFormat.Quality,
		humanize.Bytes(uint64(fileSize)),
	), &tb.SendOptions{
		ParseMode:             tb.ModeHTML,
		ReplyMarkup:           r,
		DisableWebPagePreview: true,
	},
	)
	return nil
}

// Streams the video directly to Google Drive
func downloadVideoOnly(video *youtube.Video, videoFormat *youtube.Format, c tb.Context, m *tb.Message) (string, int64, error) {
	fileName := fmt.Sprintf("%s.mp4", video.Title)
	fileName = strings.ReplaceAll(fileName, "/", "_")
	videoStream, videoSize, err := client.GetStream(video, videoFormat)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return "", 0, err
	}

	if videoSize == 0 {
		videoSize, err = getStreamSize(video, videoFormat)
		if err != nil {
			logger.ErrorLogger.Println(err)
			return "", 0, err
		}
	}

	upload, err := uploadVideo(videoStream, fileName, video, videoFormat, videoSize, c, m)

	return upload, videoSize, err
}

// Downloads the video and audio to a file and uploads it to Google Drive
func downloadVideoAudio(video *youtube.Video, videoFormat *youtube.Format, audioFormat *youtube.Format, c tb.Context, m *tb.Message, fileName string) (string, int64, error) {

	videoStream, videoSize, err := client.GetStream(video, videoFormat)
	if err != nil {
		return "", 0, err
	}

	if videoSize == 0 {
		videoSize, err = getStreamSize(video, videoFormat)
		if err != nil {
			return "", 0, err
		}
	}

	audioStream, audioSize, err := client.GetStream(video, audioFormat)
	if err != nil {
		return "", 0, err
	}

	if audioSize == 0 {
		audioSize, err = getStreamSize(video, audioFormat)
		if err != nil {
			return "", 0, err
		}
	}

	videoFile, err := os.Create("./downloads/" + video.ID + ".m4v")
	if err != nil {
		return "", 0, err
	}

	audioFile, err := os.Create("./downloads/" + video.ID + ".m4a")
	if err != nil {
		return "", 0, err
	}

	videoPassThru := &PassThru{ReadCloser: videoStream}
	defer videoPassThru.Close()
	audioPassThru := &PassThru{ReadCloser: audioStream}
	defer audioPassThru.Close()

	videoDownloaded := false
	audioDownloaded := false

	go func() {
		io.Copy(videoFile, videoPassThru)
		videoDownloaded = true
		videoFile.Close()
	}()

	go func() {
		io.Copy(audioFile, audioPassThru)
		audioDownloaded = true
		audioFile.Close()
	}()

	now := time.Now()

	for {
		if videoDownloaded && audioDownloaded {
			break
		}

		current := videoPassThru.current + audioPassThru.current
		total := videoSize + audioSize

		c.Bot().Edit(m, fmt.Sprintf("*%s* \n_Downloading_ %.2f%% %s of %s done. \n\n⬇️ %s/s",
			video.Title,
			float64(current)/float64(total)*100,
			humanize.Bytes(uint64(current)),
			humanize.Bytes(uint64(total)),
			humanize.Bytes(uint64(float64(current)/time.Since(now).Seconds())),
		), &tb.SendOptions{
			ParseMode: tb.ModeMarkdown,
		},
		)

		time.Sleep(time.Second)
	}

	dest := fmt.Sprintf("./downloads/%s.mp4", video.ID)

	// Post-processing
	c.Bot().Edit(m, fmt.Sprintf("*%s* \n_Post-processing_", video.Title), &tb.SendOptions{
		ParseMode: tb.ModeMarkdown,
	})

	ffmpegVersionCmd := exec.Command("ffmpeg", "-y",
		"-i", videoFile.Name(),
		"-i", audioFile.Name(),
		"-c", "copy",
		"-shortest",
		dest,
		"-loglevel", "warning",
	)
	ffmpegVersionCmd.Stderr = os.Stderr
	ffmpegVersionCmd.Stdout = os.Stdout

	err = ffmpegVersionCmd.Run()

	if err != nil {
		return "", 0, err
	}

	os.Remove(videoFile.Name())
	os.Remove(audioFile.Name())

	destFile, err := os.Open(dest)
	if err != nil {
		return "", 0, err
	}

	fileStats, _ := destFile.Stat()
	fileSize := fileStats.Size()

	driveLink, err := uploadVideo(destFile, fileName, video, videoFormat, fileSize, c, m)
	destFile.Close()

	os.Remove(dest)

	return driveLink, fileSize, err
}

func uploadVideo(r io.Reader, fileName string, video *youtube.Video, videoFormat *youtube.Format, size int64, c tb.Context, m *tb.Message) (string, error) {
	uploadFinished := false
	uploadCurrent := int64(0)
	now := time.Now()

	go func() {
		for {
			if uploadFinished {
				break
			}

			c.Bot().Edit(m, fmt.Sprintf("*%s* _%s_]\n_Uploading_ (%.2f %%) %s of %s\n\n⬆️ %s/s",
				video.Title,
				videoFormat.Quality,
				float64(uploadCurrent)/float64(size)*100,
				humanize.Bytes(uint64(uploadCurrent)),
				humanize.Bytes(uint64(size)),
				humanize.Bytes(uint64(float64(uploadCurrent)/time.Since(now).Seconds())),
			), &tb.SendOptions{
				ParseMode: tb.ModeMarkdown,
			},
			)

			time.Sleep(time.Second)
		}
	}()

	upload, err := gdrive.DriveClient.UploadReader(r, fileName, configs.GetTeamDriveFolderID(), func(c, _ int64) {
		uploadCurrent = c
		// total always returns 0
	})

	uploadFinished = true

	return upload, err
}

func getStreamSize(video *youtube.Video, format *youtube.Format) (int64, error) {
	url, err := client.GetStreamURL(video, format)
	if err != nil {
		return 0, err
	}

	resp, err := httpClient.Head(url)
	return resp.ContentLength, err
}

func genFilename(name string) string {
	fileName := fmt.Sprintf("%s.mp4", name)
	return strings.ReplaceAll(fileName, "/", "_")
}
