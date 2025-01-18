package handlers

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetFileHandler(c echo.Context) error {
	filename := c.Param("filename")
	filePath := filepath.Join("./downloads", filename)

	if !fileExists(filePath) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found"})
	}

	file, err := os.Open(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot open file"})
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot stat file"})
	}
	fileSize := fileInfo.Size()

	rangeHeader := c.Request().Header.Get("Range")
	if rangeHeader == "" {
		c.Response().Header().Set("Content-Type", "video/mp4")
		c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		_, err = io.Copy(c.Response(), file)
		return err
	}

	start, end, err := parseRange(rangeHeader, fileSize)
	if err != nil {
		return c.JSON(http.StatusRequestedRangeNotSatisfiable, map[string]string{"error": "invalid range"})
	}

	c.Response().Header().Set("Content-Type", "video/mp4")
	c.Response().Header().Set("Accept-Ranges", "bytes")
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	c.Response().Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Response().WriteHeader(http.StatusPartialContent)

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot seek file"})
	}

	_, err = io.CopyN(c.Response(), file, end-start+1)
	return err
}

func parseRange(rangeHeader string, fileSize int64) (int64, int64, error) {
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		return 0, 0, errors.New("invalid range format")
	}

	rangeHeader = strings.TrimPrefix(rangeHeader, prefix)
	parts := strings.Split(rangeHeader, "-")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range format")
	}

	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid range start")
	}

	var end int64
	if parts[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid range end")
		}
	}

	if start > end || start < 0 || end >= fileSize {
		return 0, 0, errors.New("invalid range values")
	}

	return start, end, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
