package models

type File struct {
	ID             uint   `gorm:"primaryKey"`
	TorrentID      uint   `gorm:"index"`
	Name           string `gorm:"type:text"`
	Size           int64
	DownloadedSize int64
	Status         string `gorm:"type:varchar(20);default:in_progress"`
	Path           string `gorm:"type:text"`
}
