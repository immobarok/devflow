package websocket

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jwtpkg "github.com/mobarok/devflow/pkg/jwt"
	"github.com/mobarok/devflow/pkg/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub       *Hub
	jwtSecret string
}

func NewHandler(hub *Hub, jwtSecret string) *Handler {
	return &Handler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

// HandleWebSocket upgrades the HTTP connection to WebSocket and registers the client.
func (h *Handler) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token = parts[1]
			}
		}
	}

	if token == "" {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token is required for WebSocket connection")
		return
	}

	claims, err := jwtpkg.ValidateToken(token, h.jwtSecret)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		Hub:    h.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: claims.UserID,
	}

	h.hub.Register <- client

	go client.WritePump()
	client.ReadPump()
}
