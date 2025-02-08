package handlers

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func StreamTorrentHandler(c echo.Context) error {
	filepathParam := c.Param("filepath")

	torrentFilePath := filepath.Join("uploads", filepathParam+".torrent")
	log.Printf("Trying to load torrent from file: %s", torrentFilePath)

	t, err := client.AddTorrentFromFile(torrentFilePath)
	if err != nil {
		log.Printf("Failed to load torrent from file: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load torrent"})
	}
	if t.Stats().TotalPeers == 0 {
		log.Println("No peers connected, waiting for peers to join...")
		time.Sleep(5 * time.Second)
	}

	log.Printf("Torrent added: %s", filepathParam)
	<-t.GotInfo()
	log.Printf("Torrent metadata loaded: %s", filepathParam)

	log.Println("Files in torrent:")
	for _, f := range t.Files() {
		log.Printf("- %s", f.Path())
	}

	requestedFilePath := c.QueryParam("file")
	if requestedFilePath == "" {
		log.Println("No file specified in query parameters")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file parameter is required"})
	}
	log.Printf("Searching for file: %s", requestedFilePath)

	var file *torrent.File
	for _, f := range t.Files() {
		if f.Path() == requestedFilePath {
			file = f
			break
		}
	}

	if file == nil {
		log.Printf("File not found in torrent: %s", requestedFilePath)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	log.Printf("File found: %s, size: %d bytes", file.Path(), file.Length())
	log.Printf("Torrent status: %v, Peers: %d, Downloaded: %d bytes, Total Size: %d bytes",
		t.Stats().ConnStats, t.Stats().TotalPeers, t.Stats().PiecesComplete, t.Length())

	for _, f := range t.Files() {
		log.Printf("File: %s, Priority: %v, Bytes Completed: %d", f.Path(), f.Priority(), f.BytesCompleted())
	}

	file.SetPriority(torrent.PiecePriorityNow)

	waitForFileReady(file)

	rangeHeader := c.Request().Header.Get("Range")
	start, end, err := parseRange(rangeHeader, file.Length())
	if err != nil {
		log.Printf("Invalid Range header: %s", rangeHeader)
		return c.JSON(http.StatusRequestedRangeNotSatisfiable, map[string]string{"error": "invalid range"})
	}

	log.Printf("Streaming file: %s, range: %d-%d", requestedFilePath, start, end)

	c.Response().Header().Set("Content-Type", "video/mp4")
	c.Response().Header().Set("Accept-Ranges", "bytes")
	c.Response().Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.Length()))
	c.Response().Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	c.Response().WriteHeader(http.StatusPartialContent)

	reader := file.NewReader()
	_, err = reader.Seek(start, io.SeekStart)
	if err != nil {
		log.Printf("Failed to seek file %s: %v", requestedFilePath, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to seek file"})
	}

	done := c.Request().Context().Done()
	copyCh := make(chan int64)

	go func() {
		n, err := io.CopyN(c.Response(), reader, end-start+1)
		if err != nil && err != io.EOF {
			log.Printf("Error during streaming %s: %v", requestedFilePath, err)
		} else if err == io.EOF {
			log.Printf("File %s streamed completely", requestedFilePath)
		} else {
			log.Printf("Bytes streamed: %d", n)
		}
		copyCh <- n
	}()

	select {
	case <-done:
		log.Println("Client disconnected")
		return nil
	case n := <-copyCh:
		log.Printf("Bytes streamed: %d", n)
	}

	return nil
}

func waitForFileReady(file *torrent.File) {
	for {
		bytesCompleted := file.BytesCompleted()
		fileSize := file.Length()

		if fileSize == 0 {
			log.Printf("File size is zero, waiting...")
		} else {
			progress := float64(bytesCompleted) / float64(fileSize)
			log.Printf("Waiting for file to buffer... Progress: %.2f%%", progress*100)
			if progress > 0.1 {
				break
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	if rangeHeader == "" {
		return 0, fileSize - 1, nil
	}

	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return 0, fileSize - 1, fmt.Errorf("invalid range format")
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, prefix)
	parts := strings.Split(rangeHeader, "-")

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fileSize - 1, fmt.Errorf("invalid range start")
	}

	var end int64 = fileSize - 1
	if len(parts) > 1 && parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, fileSize - 1, fmt.Errorf("invalid range end")
		}
	}

	if start > end || start < 0 || end >= fileSize {
		return 0, fileSize - 1, fmt.Errorf("range out of bounds")
	}

	return start, end, nil
}
