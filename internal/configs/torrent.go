package configs

import (
	"math/rand"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/storage"
	"pupadrive.go/internal/logger"
)

func GetTorrentConfig() *torrent.ClientConfig {
	blocklist, err := iplist.MMapPackedFile("packed-blocklist")
	if err != nil {
		logger.ErrorLogger.Println(err)
	}
	rand.Seed(time.Now().Unix())
	min := 10000
	max := 44444
	random := rand.Intn(max-min) + min

	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.HTTPProxy = GetProxy()
	clientConfig.HTTPUserAgent = "Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.0; T312461; .NET CLR 1.0.3705)"
	clientConfig.Seed = false
	clientConfig.ListenPort = random
	clientConfig.DisableIPv6 = false
	clientConfig.AcceptPeerConnections = true
	clientConfig.DefaultStorage = storage.NewFileByInfoHash("downloads")
	clientConfig.AlwaysWantConns = false
	clientConfig.IPBlocklist = blocklist

	return clientConfig
}
