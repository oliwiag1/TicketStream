package handlers

import (
	"errors"
	"log"
	"net/http"

	"ticketstream/backend/internal/services"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type WSHandler struct {
	logger   *log.Logger
	realtime *services.RealtimeService
}

func NewWSHandler(logger *log.Logger, realtime *services.RealtimeService) *WSHandler {
	return &WSHandler{logger: logger, realtime: realtime}
}

func (h *WSHandler) Connect(c echo.Context) error {
	eventID := c.Param("eventId")

	exists, err := h.realtime.EventExists(c.Request().Context(), eventID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "event lookup failed"})
	}
	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "event_not_found"})
	}

	handler := websocket.Handler(func(conn *websocket.Conn) {
		defer conn.Close()
		if err := h.realtime.ServeConnection(c.Request().Context(), eventID, conn); err != nil && !errors.Is(err, services.ErrEventNotFound) {
			h.logger.Printf("websocket connection closed with error event=%s err=%v", eventID, err)
		}
	})

	handler.ServeHTTP(c.Response(), c.Request())
	return nil
}
