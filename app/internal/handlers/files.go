package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"torrent-getter/internal/services"
)

func ListFilesHandler(c echo.Context) error {
	status := c.QueryParam("status")
	files, err := services.GetFiles(status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, files)
}
