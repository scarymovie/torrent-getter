package models

import "time"

type Torrent struct {
	ID         uint   `gorm:"primaryKey"`
	MagnetLink string `gorm:"type:text"`
	InfoHash   string `gorm:"uniqueIndex"`
	Status     string `gorm:"type:varchar(20);default:in_progress"`
	CreatedAt  time.Time
}
