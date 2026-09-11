/*
 * 即時聊天 WebSocket 介面。
 * 1. 簽發短時 WebSocket ticket 並建立單會話連線。
 * 2. 重用 ChatService 的消息與大廈權限校驗。
 * 3. 以本地連線中心和可選 Redis 廣播分發已提交的消息事件。
 */
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
	"ajoliving_web/http_service/internal/utils"
)

const (
	realtimeWriteWait  = 10 * time.Second
	realtimePongWait   = 60 * time.Second
	realtimePingPeriod = (realtimePongWait * 9) / 10
	realtimeMaxMessage = 32 * 1024
)

// 1. RealtimeHandler handles ticket issuance and WebSocket connections.
type RealtimeHandler struct {
	authService *service.AuthService
	chatService *service.ChatService
	hub         *RealtimeHub
	upgrader    websocket.Upgrader
}

// 2. RealtimeHub keeps local chat connections and forwards cross-instance events.
type RealtimeHub struct {
	mu            sync.RWMutex
	clients       map[string]map[*realtimeClient]struct{}
	broker        service.RealtimeBroker
	chatService   *service.ChatService
	instance      string
	consumeCtx    context.Context
	consumeCancel context.CancelFunc
}

// 3. realtimeClient stores one authenticated chat connection.
type realtimeClient struct {
	conn   *websocket.Conn
	chatID string
	userID int64
	send   chan []byte
	done   chan struct{}
}

// 4. realtimeCommand defines the small JSON commands accepted by WebSocket.
type realtimeCommand struct {
	Type            string   `json:"type"`
	ChatID          string   `json:"chat_id"`
	Content         string   `json:"content"`
	AttachmentIDs   []string `json:"attachment_ids,omitempty"`
	ClientMessageID string   `json:"client_message_id,omitempty"`
}

// 5. realtimeMessageEvent defines one broadcast message event.
type realtimeMessageEvent struct {
	Type            string                  `json:"type"`
	ChatID          string                  `json:"chat_id"`
	Message         service.MessageResponse `json:"message"`
	ClientMessageID string                  `json:"client_message_id,omitempty"`
	Origin          string                  `json:"origin,omitempty"`
}

// 6. NewRealtimeHub creates a local hub and starts the optional broker consumer.
func NewRealtimeHub(broker service.RealtimeBroker, chatService *service.ChatService) *RealtimeHub {
	consumeCtx, consumeCancel := context.WithCancel(context.Background())
	hub := &RealtimeHub{
		clients:       make(map[string]map[*realtimeClient]struct{}),
		broker:        broker,
		chatService:   chatService,
		instance:      utils.NewPublicID(),
		consumeCtx:    consumeCtx,
		consumeCancel: consumeCancel,
	}
	if broker != nil {
		go hub.consumeBroker()
	}
	return hub
}

// 6.1 Close stops the broker consumer owned by this hub.
func (h *RealtimeHub) Close() {
	if h != nil && h.consumeCancel != nil {
		h.consumeCancel()
	}
}

// 7. NewRealtimeHandler creates the realtime HTTP handler.
func NewRealtimeHandler(authService *service.AuthService, chatService *service.ChatService, hub *RealtimeHub, cfg *config.Config) *RealtimeHandler {
	return &RealtimeHandler{
		authService: authService,
		chatService: chatService,
		hub:         hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(request *http.Request) bool {
				return isAllowedRealtimeOrigin(request, cfg)
			},
		},
	}
}

// 8. IssueTicket returns a short-lived ticket for a WebSocket handshake.
func (h *RealtimeHandler) IssueTicket(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	ticket, expiresIn, err := h.authService.IssueRealtimeTicket(user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"ticket": ticket, "expires_in": expiresIn})
}

// 9. HandleWebSocket authenticates and serves one chat-scoped connection.
func (h *RealtimeHandler) HandleWebSocket(c *gin.Context) {
	ticket := strings.TrimSpace(c.Query("ticket"))
	chatID := strings.TrimSpace(c.Query("chat_id"))
	if ticket == "" || chatID == "" {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "realtime ticket and chat_id are required"))
		return
	}
	identity, err := h.authService.AuthenticateRealtimeToken(c.Request.Context(), ticket)
	if err != nil || identity.MemberStatus != "active" {
		if err == nil {
			err = errcode.New(errcode.CodeAuthForbidden, "active membership is required")
		}
		errcode.WriteError(c, err)
		return
	}
	if err := h.chatService.ValidateChatAccess(c.Request.Context(), identity.UserID, chatID); err != nil {
		errcode.WriteError(c, err)
		return
	}

	connection, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := &realtimeClient{conn: connection, chatID: chatID, userID: identity.UserID, send: make(chan []byte, 32), done: make(chan struct{})}
	h.hub.register(client)
	defer func() {
		h.hub.unregister(client)
		_ = connection.Close()
	}()

	go h.writePump(client)
	connection.SetReadLimit(realtimeMaxMessage)
	_ = connection.SetReadDeadline(time.Now().Add(realtimePongWait))
	connection.SetPongHandler(func(string) error {
		return connection.SetReadDeadline(time.Now().Add(realtimePongWait))
	})
	for {
		_, payload, readErr := connection.ReadMessage()
		if readErr != nil {
			break
		}
		var command realtimeCommand
		if err := json.Unmarshal(payload, &command); err != nil {
			h.sendError(client, errcode.New(errcode.CodeValidationError, "invalid realtime command"))
			continue
		}
		if strings.TrimSpace(command.Type) != "message" || strings.TrimSpace(command.ChatID) != chatID {
			h.sendError(client, errcode.New(errcode.CodeValidationError, "unsupported realtime command"))
			continue
		}
		message, sendErr := h.chatService.SendRealtimeMessageWithAttachmentsAndClientID(c.Request.Context(), identity.UserID, chatID, command.Content, command.AttachmentIDs, command.ClientMessageID)
		if sendErr != nil {
			h.sendError(client, sendErr)
			continue
		}
		h.hub.PublishMessage(c.Request.Context(), chatID, *message, command.ClientMessageID)
	}
}

// 10. PublishMessage broadcasts a committed message to local and remote clients.
func (h *RealtimeHub) PublishMessage(ctx context.Context, chatID string, message service.MessageResponse, clientMessageID string) {
	event := realtimeMessageEvent{Type: "message", ChatID: strings.TrimSpace(chatID), Message: message, ClientMessageID: strings.TrimSpace(clientMessageID), Origin: h.instance}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.broadcast(ctx, event.ChatID, payload)
	if h.broker != nil {
		if err := h.broker.Publish(ctx, payload); err == nil && h.chatService != nil {
			h.chatService.MarkRealtimeEventPublished(ctx, message.MessageID)
		}
	}
}

// 11. register adds one connection to its chat room.
func (h *RealtimeHub) register(client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.chatID] == nil {
		h.clients[client.chatID] = make(map[*realtimeClient]struct{})
	}
	h.clients[client.chatID][client] = struct{}{}
}

// 12. unregister removes one connection from its chat room.
func (h *RealtimeHub) unregister(client *realtimeClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.clients[client.chatID]
	delete(clients, client)
	select {
	case <-client.done:
	default:
		close(client.done)
	}
	if len(clients) == 0 {
		delete(h.clients, client.chatID)
	}
}

// 13. broadcast sends one event to every local chat connection.
func (h *RealtimeHub) broadcast(ctx context.Context, chatID string, payload []byte) {
	h.mu.RLock()
	clients := make([]*realtimeClient, 0, len(h.clients[chatID]))
	for client := range h.clients[chatID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		if h.chatService != nil {
			if err := h.chatService.ValidateChatAccess(ctx, client.userID, chatID); err != nil {
				_ = client.conn.Close()
				continue
			}
		}
		select {
		case <-client.done:
			continue
		case client.send <- payload:
		default:
			if client.conn != nil {
				_ = client.conn.Close()
			}
		}
	}
}

// 14. consumeBroker forwards events received from other Go instances.
func (h *RealtimeHub) consumeBroker() {
	if h == nil {
		return
	}
	ctx := h.consumeCtx
	if ctx == nil {
		ctx = context.Background()
	}
	h.consumeBrokerContext(ctx)
}

// 14.1 consumeBrokerContext reconnects after a Redis subscription failure or disconnect.
func (h *RealtimeHub) consumeBrokerContext(ctx context.Context) {
	if h == nil || h.broker == nil {
		return
	}
	const (
		initialRetryDelay = 100 * time.Millisecond
		maxRetryDelay     = 5 * time.Second
	)
	retryDelay := initialRetryDelay
	for {
		if ctx.Err() != nil {
			return
		}
		channel, closeSubscription, err := h.broker.Subscribe(ctx)
		if err != nil {
			if !waitRealtimeBrokerRetry(ctx, retryDelay) {
				return
			}
			retryDelay *= 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
			continue
		}
		connected := true
		for connected {
			select {
			case <-ctx.Done():
				if closeSubscription != nil {
					closeSubscription()
				}
				return
			case payload, ok := <-channel:
				if !ok {
					connected = false
					continue
				}
				retryDelay = initialRetryDelay
				var event realtimeMessageEvent
				if json.Unmarshal(payload, &event) != nil || event.Origin == h.instance || event.ChatID == "" {
					continue
				}
				h.broadcast(ctx, event.ChatID, payload)
			}
		}
		if closeSubscription != nil {
			closeSubscription()
		}
		if !waitRealtimeBrokerRetry(ctx, retryDelay) {
			return
		}
		retryDelay *= 2
		if retryDelay > maxRetryDelay {
			retryDelay = maxRetryDelay
		}
	}
}

// 14.2 waitRealtimeBrokerRetry waits without blocking hub shutdown.
func waitRealtimeBrokerRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// 15. writePump writes queued events and heartbeat pings.
func (h *RealtimeHandler) writePump(client *realtimeClient) {
	ticker := time.NewTicker(realtimePingPeriod)
	defer ticker.Stop()
	for {
		select {
		case payload, ok := <-client.send:
			if !ok {
				return
			}
			_ = client.conn.SetWriteDeadline(time.Now().Add(realtimeWriteWait))
			if client.conn.WriteMessage(websocket.TextMessage, payload) != nil {
				return
			}
		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(realtimeWriteWait))
			if client.conn.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		case <-client.done:
			return
		}
	}
}

// 16. sendError queues a stable error event without writing from the read loop.
func (h *RealtimeHandler) sendError(client *realtimeClient, err error) {
	code, message := errcode.CodeInternalError, "internal server error"
	var appErr *errcode.AppError
	if errors.As(err, &appErr) {
		code, message = appErr.Code, appErr.Message
	}
	payload, marshalErr := json.Marshal(map[string]string{"type": "error", "code": code, "message": message})
	if marshalErr != nil {
		return
	}
	select {
	case <-client.done:
		return
	case client.send <- payload:
	default:
	}
}

// 17. isAllowedRealtimeOrigin checks same-host and configured CORS origins.
func isAllowedRealtimeOrigin(request *http.Request, cfg *config.Config) bool {
	origin := strings.TrimSpace(request.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if strings.EqualFold(parsed.Host, request.Host) {
		return true
	}
	if cfg == nil {
		return false
	}
	for _, allowed := range strings.Split(cfg.CORSAllowedOrigins, ",") {
		if strings.EqualFold(strings.TrimRight(strings.TrimSpace(allowed), "/"), strings.TrimRight(origin, "/")) {
			return true
		}
	}
	return false
}
