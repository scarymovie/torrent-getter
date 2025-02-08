package services

import (
	"github.com/anacrolix/torrent"
)

func ProcessTorrentFromFile(filePath string, client *torrent.Client) (string, error) {
	t, err := client.AddTorrentFromFile(filePath)
	if err != nil {
		return "", err
	}

	<-t.GotInfo()
	t.DownloadAll()

	return t.InfoHash().HexString(), nil
}
