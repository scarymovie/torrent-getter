package services

import (
	"errors"
	"github.com/anacrolix/torrent"
	"path/filepath"
	"torrent-getter/internal/models"
	"torrent-getter/internal/repositories"
)

func ProcessTorrentFile(path string, client *torrent.Client) (string, error) {
	if client == nil {
		return "", errors.New("torrent client is not initialized")
	}

	t, err := client.AddTorrentFromFile(path)
	if err != nil {
		return "", err
	}

	<-t.GotInfo()

	infoHash := t.InfoHash().HexString()
	newTorrent, err := repositories.CreateTorrent(&models.Torrent{
		InfoHash: infoHash,
		Status:   "in_progress",
	})
	if err != nil {
		return "", err
	}

	for _, file := range t.Files() {
		repositories.CreateFile(&models.File{
			TorrentID:      newTorrent.ID,
			Name:           file.Path(),
			Size:           file.Length(),
			DownloadedSize: 0,
			Status:         "in_progress",
			Path:           filepath.Join("./downloads", file.Path()),
		})
	}

	t.DownloadAll()
	return infoHash, nil
}
