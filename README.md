<img align="right" src="https://user-images.githubusercontent.com/36324542/148573650-9244c1a3-406c-411f-aa4a-3550aa7bafb9.png" height="300px">

# Pupadrive

A simple bot for downloading torrents and youtube videos to your google drive.

This bot is meant to be used to download CC0 licenced or any other free content, we do not support nor recommend using it for illegal activities.

## Installation

### Setting up config file

Fill up rest of the fields. Meaning of each fields are described below:

- **PROXY**: Proxy for torrent client. socks5 and http proxies are supported.
- **HTTP_PROXY**: Proxy for youtube client. Only http proxies are supported.
- **BOT_TOKEN**: Telegram bot token
- **DRIVE_ID**: GDrive ID/ shared drive id
- **GOOGLE_DRIVE_FOLDER**: GDrive folder id for uploading youtube videos

### Install and run

Please ensure you have installed Go 1.17 or later.

```shell
git clone https://github.com/pupagang/pupadrive-go
go get && go build ./cmd/pupadrive/main.go
./main
```
