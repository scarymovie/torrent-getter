package handlers

import (
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"torrent-getter/internal/services"
)

type TorrentUploadRequest struct {
	MagnetLink string `json:"magnet_link,omitempty"`
}

var client *torrent.Client

func SetTorrentClient(c *torrent.Client) {
	client = c
}

func UploadTorrentHandler(c echo.Context) error {
	var req TorrentUploadRequest

	torrentFile, err := c.FormFile("torrent_file")
	if err == nil {
		file, err := torrentFile.Open()
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid file"})
		}
		defer file.Close()

		infoHash, err := services.ProcessTorrentFromReader(file, client)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		log.Printf("Torrent added with InfoHash: %s", infoHash)

		return c.JSON(http.StatusOK, map[string]string{"info_hash": infoHash, "status": "download started"})
	}

	if err := c.Bind(&req); err != nil || req.MagnetLink == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
	}

	infoHash, err := services.ProcessMagnetLink(req.MagnetLink, client)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("Magnet link added with InfoHash: %s", infoHash)

	return c.JSON(http.StatusOK, map[string]string{"info_hash": infoHash, "status": "download started"})
}
