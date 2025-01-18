package main

import (
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
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

	go func() {
		log.Println("Starting MonitorDownloads...")
		MonitorDownloads(client)
	}()

	handlers.SetTorrentClient(client)

	e := echo.New()

	e.POST("/upload", handlers.UploadTorrentHandler)
	e.GET("/files", handlers.ListFilesHandler)
	e.GET("/files/:filename", handlers.GetFileHandler)

	log.Println("Starting HTTP server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}

func MonitorDownloads(client *torrent.Client) {
	for {
		torrents := client.Torrents()
		if len(torrents) == 0 {
			log.Println("No torrents in the client")
		} else {
			for _, t := range torrents {
				log.Printf("Monitoring torrent: %s, ActivePeers: %d",
					t.Name(), t.Stats().ActivePeers)

				allFilesCompleted := true

				for _, file := range t.Files() {
					progress := file.BytesCompleted()
					totalSize := file.Length()

					log.Printf("File: %s, Progress: %d / %d bytes",
						file.Path(), progress, totalSize)

					err := repositories.UpdateFileProgress(file.Path(), progress)
					if err != nil {
						log.Printf("Failed to update progress for file %s: %v", file.Path(), err)
					}

					if progress < totalSize {
						allFilesCompleted = false
					}
				}

				if allFilesCompleted {
					log.Printf("Torrent %s fully downloaded, stopping download", t.Name())
					err := repositories.UpdateTorrentStatus(t.InfoHash().HexString(), "completed")
					if err != nil {
						log.Printf("Failed to update torrent status: %v", err)
					}

					t.Drop()
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}
