package handlers

import (
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"torrent-getter/internal/services"
)

type TorrentUploadRequest struct {
	FilePath string `json:"file_path"`
}

var client *torrent.Client

func SetTorrentClient(c *torrent.Client) {
	client = c
}

func UploadTorrentHandler(c echo.Context) error {
	var req TorrentUploadRequest

	if err := c.Bind(&req); err != nil || req.FilePath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input, file_path is required"})
	}

	infoHash, err := services.ProcessTorrentFromFile(req.FilePath, client)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("Torrent file added from path %s with InfoHash: %s", req.FilePath, infoHash)
	return c.JSON(http.StatusOK, map[string]string{"info_hash": infoHash, "status": "download started"})
}
