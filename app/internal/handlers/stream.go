package handlers

import (
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func StreamTorrentHandler(c echo.Context) error {
	infoHash := c.Param("infoHash")
	filename := c.Param("filename")

	hash := metainfo.NewHashFromHex(infoHash)

	t, _ := client.Torrent(hash)
	if t == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "torrent not found"})
	}

	<-t.GotInfo()

	var file *torrent.File
	found := false
	for _, f := range t.Files() {
		if f.Path() == filename {
			file = f
			found = true
			break
		}
	}

	if !found {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	rangeHeader := c.Request().Header.Get("Range")
	var start, end int64
	if rangeHeader != "" {
		start, end, _ = parseRange(rangeHeader, file.Length())
	} else {
		start = 0
		end = file.Length() - 1
	}

	c.Response().Header().Set("Content-Type", "video/mp4")
	c.Response().Header().Set("Accept-Ranges", "bytes")
	c.Response().Header().Set("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(file.Length(), 10))
	c.Response().Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	c.Response().WriteHeader(http.StatusPartialContent)

	reader := file.NewReader()
	_, err := reader.Seek(start, io.SeekStart)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to seek file"})
	}
	defer func() {
		log.Printf("Deleting cached file: %s", file.Path())
		err := os.Remove(filepath.Join("./downloads", file.Path()))
		if err != nil {
			log.Printf("Failed to delete file: %v", err)
		}
	}()

	_, err = reader.Seek(start, io.SeekStart)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to seek file"})
	}

	_, err = io.CopyN(c.Response(), reader, end-start+1)
	if err != nil && err != io.EOF {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to stream file"})
	}

	return nil
}

func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return 0, 0, nil
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, prefix)
	parts := strings.Split(rangeHeader, "-")
	start, _ := strconv.ParseInt(parts[0], 10, 64)

	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, _ = strconv.ParseInt(parts[1], 10, 64)
	}

	return start, end, nil
}
