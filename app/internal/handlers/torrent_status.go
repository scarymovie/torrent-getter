package handlers

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/labstack/echo/v4"
	"net/http"
)

func TorrentStatusHandler(c echo.Context) error {
	infoHash := c.Param("infoHash")

	t, _ := client.Torrent(metainfo.NewHashFromHex(infoHash))
	if t == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "torrent not found"})
	}

	<-t.GotInfo()

	files := t.Files()
	fileStatuses := make([]map[string]interface{}, len(files))

	for i, file := range files {
		fileStatuses[i] = map[string]interface{}{
			"name":       file.Path(),
			"size":       file.Length(),
			"downloaded": file.BytesCompleted(),
		}
	}

	downloadSpeed := calculateDownloadSpeed(t)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"info_hash":         infoHash,
		"name":              t.Name(),
		"total_peers":       t.Stats().TotalPeers,
		"active_peers":      t.Stats().ActivePeers,
		"connected_seeders": t.Stats().ConnectedSeeders,
		"pieces_complete":   t.Stats().PiecesComplete,
		"download_speed":    downloadSpeed,
		"files":             fileStatuses,
	})

}

var previousProgress = make(map[string]int64)

func calculateDownloadSpeed(t *torrent.Torrent) float64 {
	currentProgress := int64(0)
	for _, file := range t.Files() {
		currentProgress += file.BytesCompleted()
	}

	lastProgress := previousProgress[t.InfoHash().HexString()]
	speed := float64(currentProgress-lastProgress) / 5.0

	previousProgress[t.InfoHash().HexString()] = currentProgress

	return speed
}
