package lcu

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	readyCheckEvent  = "OnJsonApiEvent_lol-matchmaking_v1_ready-check"
	champSelectEvent = "OnJsonApiEvent_lol-champ-select_v1_session"
)

type AutoPickSettings struct {
	Enabled        bool
	PrimaryChamp   int
	SecondaryChamp int
	ChampPool      []int
}

type AutomationSettings struct {
	AutoAccept bool
	AutoPick   AutoPickSettings
}

type automationState struct {
	handledActions map[string]struct{}
}

type eventEnvelope struct {
	Data      json.RawMessage `json:"data"`
	EventType string          `json:"eventType"`
}

type readyCheckData struct {
	State          string `json:"state"`
	PlayerResponse string `json:"playerResponse"`
}

type champSelectSession struct {
	GameID int64 `json:"gameId"`
	Timer  struct {
		Phase string `json:"phase"`
	} `json:"timer"`
	LocalPlayerCellID int `json:"localPlayerCellId"`
	Actions           [][]struct {
		ID           int    `json:"id"`
		ActorCellID  int    `json:"actorCellId"`
		Type         string `json:"type"`
		Completed    bool   `json:"completed"`
		IsInProgress bool   `json:"isInProgress"`
	} `json:"actions"`
}

// StartWebSocket bağlantı kopana kadar eventleri dinler ve geçici hatalarda
// context iptal edilene kadar artan gecikmeyle yeniden bağlanır.
func (c *Client) StartWebSocket(ctx context.Context, settings func() AutomationSettings) {
	backoff := time.Second
	for {
		connectedAt := time.Now()
		err := c.listenWebSocket(ctx, settings)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Println("LCU WebSocket bağlantısı sona erdi:", err)
		}

		if time.Since(connectedAt) > 30*time.Second {
			backoff = time.Second
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}

		if backoff < 30*time.Second {
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}
}

func (c *Client) listenWebSocket(ctx context.Context, settings func() AutomationSettings) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: lcuRequestTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}
	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.Token))
	headers := http.Header{"Authorization": []string{"Basic " + auth}}
	url := fmt.Sprintf("wss://127.0.0.1:%s/", c.Port)

	conn, _, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		return fmt.Errorf("bağlantı kurulamadı: %w", err)
	}
	defer conn.Close()

	stopCloseWatcher := make(chan struct{})
	defer close(stopCloseWatcher)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-stopCloseWatcher:
		}
	}()

	for _, eventName := range []string{readyCheckEvent, champSelectEvent} {
		subscription, err := json.Marshal([]any{5, eventName})
		if err != nil {
			return fmt.Errorf("%s aboneliği oluşturulamadı: %w", eventName, err)
		}
		if err := conn.WriteMessage(websocket.TextMessage, subscription); err != nil {
			return fmt.Errorf("%s aboneliği gönderilemedi: %w", eventName, err)
		}
	}

	log.Println("LCU WebSocket bağlantısı kuruldu; otomasyon eventleri dinleniyor")
	state := &automationState{handledActions: make(map[string]struct{})}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("event okunamadı: %w", err)
		}

		if err := c.handleWAMPMessage(ctx, message, settings(), state); err != nil {
			log.Println("LCU otomasyon eventi işlenemedi:", err)
		}
	}
}

func (c *Client) handleWAMPMessage(ctx context.Context, message []byte, settings AutomationSettings, state *automationState) error {
	var parts []json.RawMessage
	if err := json.Unmarshal(message, &parts); err != nil {
		return fmt.Errorf("WAMP mesajı çözülemedi: %w", err)
	}
	if len(parts) != 3 {
		return nil
	}

	var messageType int
	if err := json.Unmarshal(parts[0], &messageType); err != nil || messageType != 8 {
		return nil
	}
	var eventName string
	if err := json.Unmarshal(parts[1], &eventName); err != nil {
		return fmt.Errorf("WAMP event adı çözülemedi: %w", err)
	}
	var envelope eventEnvelope
	if err := json.Unmarshal(parts[2], &envelope); err != nil {
		return fmt.Errorf("%s payload'u çözülemedi: %w", eventName, err)
	}

	switch eventName {
	case readyCheckEvent:
		return c.handleReadyCheck(ctx, envelope.Data, settings.AutoAccept)
	case champSelectEvent:
		if envelope.EventType == "Delete" {
			state.handledActions = make(map[string]struct{})
			return nil
		}
		return c.handleChampSelect(ctx, envelope.Data, settings.AutoPick, state)
	default:
		return nil
	}
}

func (c *Client) handleReadyCheck(ctx context.Context, data json.RawMessage, autoAccept bool) error {
	if !autoAccept || len(data) == 0 || string(data) == "null" {
		return nil
	}
	var readyCheck readyCheckData
	if err := json.Unmarshal(data, &readyCheck); err != nil {
		return fmt.Errorf("ready-check verisi çözülemedi: %w", err)
	}
	if readyCheck.State != "InProgress" || readyCheck.PlayerResponse != "None" {
		return nil
	}
	if err := c.AcceptMatch(ctx); err != nil {
		return fmt.Errorf("maç otomatik kabul edilemedi: %w", err)
	}
	log.Println("Maç otomatik kabul edildi")
	return nil
}

func (c *Client) handleChampSelect(ctx context.Context, data json.RawMessage, cfg AutoPickSettings, state *automationState) error {
	if !cfg.Enabled || len(data) == 0 || string(data) == "null" {
		return nil
	}
	var session champSelectSession
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("champ-select verisi çözülemedi: %w", err)
	}
	if session.Timer.Phase != "BAN_PICK" {
		return nil
	}

	for _, actionGroup := range session.Actions {
		for _, action := range actionGroup {
			if action.ActorCellID != session.LocalPlayerCellID || action.Type != "pick" || !action.IsInProgress || action.Completed {
				continue
			}

			actionKey := fmt.Sprintf("%d:%d", session.GameID, action.ID)
			if _, handled := state.handledActions[actionKey]; handled {
				return nil
			}

			pickable, err := c.GetPickableChampions(ctx)
			if err != nil {
				return err
			}
			championID := chooseChampion(cfg, pickable, rand.Intn)
			if championID == 0 {
				return errors.New("yapılandırılmış seçilebilir şampiyon bulunamadı")
			}
			if err := c.LockChampion(ctx, action.ID, championID); err != nil {
				return fmt.Errorf("%d ID'li şampiyon kilitlenemedi: %w", championID, err)
			}
			state.handledActions[actionKey] = struct{}{}
			log.Printf("%d ID'li şampiyon otomatik kilitlendi", championID)
			return nil
		}
	}
	return nil
}

func chooseChampion(cfg AutoPickSettings, pickable []int, randomIndex func(int) int) int {
	pickableSet := make(map[int]struct{}, len(pickable))
	for _, id := range pickable {
		pickableSet[id] = struct{}{}
	}

	if cfg.PrimaryChamp != 0 {
		if _, ok := pickableSet[cfg.PrimaryChamp]; ok {
			return cfg.PrimaryChamp
		}
	}
	if cfg.SecondaryChamp != 0 {
		if _, ok := pickableSet[cfg.SecondaryChamp]; ok {
			return cfg.SecondaryChamp
		}
	}

	validPool := make([]int, 0, len(cfg.ChampPool))
	for _, id := range cfg.ChampPool {
		if _, ok := pickableSet[id]; ok {
			validPool = append(validPool, id)
		}
	}
	if len(validPool) == 0 {
		return 0
	}
	return validPool[randomIndex(len(validPool))]
}
