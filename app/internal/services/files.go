package services

import (
	"github.com/anacrolix/torrent"
	"path/filepath"
	"torrent-getter/internal/models"
	"torrent-getter/internal/repositories"
)

func GetFiles(status string) ([]models.File, error) {
	return repositories.GetFilesByStatus(status)
}

func ProcessMagnetLink(link string) (string, error) {
	client, err := torrent.NewClient(nil)
	if err != nil {
		return "", err
	}
	defer client.Close()

	t, err := client.AddMagnet(link)
	if err != nil {
		return "", err
	}

	<-t.GotInfo()

	infoHash := t.InfoHash().HexString()
	newTorrent, err := repositories.CreateTorrent(&models.Torrent{
		MagnetLink: link,
		InfoHash:   infoHash,
		Status:     "in_progress",
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
