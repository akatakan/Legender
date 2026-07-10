package lcu

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const lcuRequestTimeout = 5 * time.Second

type Client struct {
	Port       string
	Token      string
	httpClient *http.Client
	metadataMu sync.RWMutex
	metadata   *ClientMetadata
}

func NewClient(port, token string) *Client {
	return &Client{
		Port:  port,
		Token: token,
		httpClient: &http.Client{
			Timeout: lcuRequestTimeout,
			Transport: &http.Transport{
				// LCU yalnızca localhost'ta self-signed sertifika ile hizmet verir.
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		},
	}
}

type SummonerInfo struct {
	DisplayName   string `json:"displayName"`
	SummonerLevel int    `json:"summonerLevel"`
	ProfileIconId int    `json:"profileIconId"`
}

type ClientMetadata struct {
	Locale      string `json:"locale"`
	Region      string `json:"region"`
	GameVersion string `json:"gameVersion"`
}

func (c *Client) GetClientMetadata(ctx context.Context) (*ClientMetadata, error) {
	c.metadataMu.RLock()
	if c.metadata != nil {
		cached := *c.metadata
		c.metadataMu.RUnlock()
		return &cached, nil
	}
	c.metadataMu.RUnlock()

	var regionLocale struct {
		Locale string `json:"locale"`
		Region string `json:"region"`
	}
	if err := c.getJSON(ctx, "/riotclient/region-locale", &regionLocale); err != nil {
		return nil, fmt.Errorf("istemci locale bilgisi alınamadı: %w", err)
	}
	var gameVersion string
	if err := c.getJSON(ctx, "/lol-patch/v1/game-version", &gameVersion); err != nil {
		return nil, fmt.Errorf("istemci oyun sürümü alınamadı: %w", err)
	}

	metadata := &ClientMetadata{Locale: regionLocale.Locale, Region: regionLocale.Region, GameVersion: gameVersion}
	c.metadataMu.Lock()
	c.metadata = metadata
	c.metadataMu.Unlock()
	result := *metadata
	return &result, nil
}

func (c *Client) GetCurrentSummoner(ctx context.Context) (*SummonerInfo, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/lol-summoner/v1/current-summoner", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("current summoner isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	if err := requireStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var summoner SummonerInfo
	if err := json.NewDecoder(resp.Body).Decode(&summoner); err != nil {
		return nil, fmt.Errorf("current summoner yanıtı çözülemedi: %w", err)
	}
	return &summoner, nil
}

func (c *Client) AcceptMatch(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/lol-matchmaking/v1/ready-check/accept", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("maç kabul isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	return requireStatus(resp, http.StatusOK, http.StatusNoContent)
}

func (c *Client) GetPickableChampions(ctx context.Context) ([]int, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/lol-champ-select/v1/pickable-champion-ids", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("seçilebilir şampiyon isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	if err := requireStatus(resp, http.StatusOK); err != nil {
		return nil, err
	}

	var champions []int
	if err := json.NewDecoder(resp.Body).Decode(&champions); err != nil {
		return nil, fmt.Errorf("seçilebilir şampiyon yanıtı çözülemedi: %w", err)
	}
	return champions, nil
}

func (c *Client) LockChampion(ctx context.Context, actionID, championID int) error {
	payload, err := json.Marshal(struct {
		ChampionID int  `json:"championId"`
		Completed  bool `json:"completed"`
	}{ChampionID: championID, Completed: true})
	if err != nil {
		return fmt.Errorf("şampiyon kilitleme isteği oluşturulamadı: %w", err)
	}

	path := fmt.Sprintf("/lol-champ-select/v1/session/actions/%d", actionID)
	req, err := c.newRequest(ctx, http.MethodPatch, path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("şampiyon kilitleme isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	return requireStatus(resp, http.StatusOK, http.StatusNoContent)
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("https://127.0.0.1:%s%s", c.Port, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("LCU isteği oluşturulamadı: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.Token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) getJSON(ctx context.Context, path string, target any) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s isteği başarısız: %w", path, err)
	}
	defer resp.Body.Close()
	if err := requireStatus(resp, http.StatusOK); err != nil {
		return err
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%s yanıtı çözülemedi: %w", path, err)
	}
	return nil
}

func requireStatus(resp *http.Response, accepted ...int) error {
	for _, status := range accepted {
		if resp.StatusCode == status {
			return nil
		}
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	detail := strings.TrimSpace(string(body))
	if detail == "" {
		return fmt.Errorf("LCU beklenmeyen durum kodu döndürdü: %d", resp.StatusCode)
	}
	return fmt.Errorf("LCU beklenmeyen durum kodu döndürdü: %d (%s)", resp.StatusCode, detail)
}
