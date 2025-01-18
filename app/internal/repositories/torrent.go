package repositories

import (
	"torrent-getter/internal/models"
)

func CreateTorrent(torrent *models.Torrent) (*models.Torrent, error) {
	return torrent, db.Create(torrent).Error
}

func UpdateTorrentStatus(infoHash string, status string) error {
	return db.Model(&models.Torrent{}).
		Where("info_hash = ?", infoHash).
		Update("status", status).Error
}
