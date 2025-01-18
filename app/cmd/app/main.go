package main

import (
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"time"
	"torrent-getter/internal/handlers"
	"torrent-getter/internal/repositories"
)

func main() {
	dsn := "host=postgres user=db_user password=db_password dbname=db_database port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}
	log.Println("Successfully connected to the database")
	repositories.SetDB(db)

	clientConfig := torrent.NewDefaultClientConfig()
	clientConfig.DataDir = "./downloads"
	clientConfig.NoDHT = false

	client, err := torrent.NewClient(clientConfig)
	if err != nil {
		log.Fatalf("Failed to initialize torrent client: %v", err)
	}
	log.Printf("Torrent client initialized with DataDir: %s", clientConfig.DataDir)
	defer client.Close()

	handlers.SetTorrentClient(client)

	CleanUpOldFiles("./downloads", 30*time.Minute)

	e := echo.New()

	e.POST("/upload", handlers.UploadTorrentHandler)
	e.GET("/files", handlers.ListFilesHandler)
	e.GET("/stream/:infoHash/:filename", handlers.StreamTorrentHandler)
	e.GET("/torrent/:infoHash/status", handlers.TorrentStatusHandler)

	log.Println("Starting HTTP server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

func CleanUpOldFiles(dataDir string, maxAge time.Duration) {
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Printf("Failed to read DataDir: %v", err)
		return
	}

	now := time.Now()
	for _, file := range files {
		filePath := filepath.Join(dataDir, file.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Failed to stat file: %v", err)
			continue
		}

		if now.Sub(info.ModTime()) > maxAge {
			err = os.Remove(filePath)
			if err != nil {
				log.Printf("Failed to delete file: %v", err)
			} else {
				log.Printf("Deleted old file: %s", filePath)
			}
		}
	}
}
