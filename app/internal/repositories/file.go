package repositories

import "torrent-getter/internal/models"

func CreateFile(file *models.File) error {
	return db.Create(file).Error
}

func UpdateFileProgress(path string, downloadedSize int64) error {
	return db.Model(&models.File{}).
		Where("path = ?", path).
		Update("downloaded_size", downloadedSize).Error
}

func GetFilesByStatus(status string) ([]models.File, error) {
	var files []models.File
	query := db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Find(&files).Error
	return files, err
}

func GetFileByName(name string) (*models.File, error) {
	var file models.File
	err := db.Where("name = ?", name).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
