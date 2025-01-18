package services

import (
	"errors"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"io"
	"path/filepath"
	"torrent-getter/internal/models"
	"torrent-getter/internal/repositories"
)

func ProcessTorrentFromReader(r io.Reader, client *torrent.Client) (string, error) {
	if client == nil {
		return "", errors.New("torrent client is not initialized")
	}

	metaInfo, err := metainfo.Load(r)
	if err != nil {
		return "", err
	}

	t, err := client.AddTorrent(metaInfo)
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
		err := repositories.CreateFile(&models.File{
			TorrentID:      newTorrent.ID,
			Name:           file.Path(),
			Size:           file.Length(),
			DownloadedSize: 0,
			Status:         "in_progress",
			Path:           filepath.Join("./downloads", file.Path()),
		})
		if err != nil {
			return "", err
		}
	}

	t.DownloadAll()

	return infoHash, nil
}
