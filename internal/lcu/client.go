package lcu

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// LCUClient HTTP isteklerini yönetecek ana yapı
type Client struct {
	Port       string
	Token      string
	httpClient *http.Client
}

// NewClient yeni bir LCU istemcisi başlatır (SSL doğrulaması kapatılmış olarak)
func NewClient(port, token string) *Client {
	return &Client{
		Port:  port,
		Token: token,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// LCU'dan dönecek Sihirdar veri modeli
type SummonerInfo struct {
	DisplayName   string `json:"displayName"`
	SummonerLevel int    `json:"summonerLevel"`
	ProfileIconId int    `json:"profileIconId"`
}

// GetCurrentSummoner giriş yapmış olan oyuncunun bilgilerini getirir
func (c *Client) GetCurrentSummoner() (*SummonerInfo, error) {
	// İstek atacağımız LCU endpoint'i
	url := fmt.Sprintf("https://127.0.0.1:%s/lol-summoner/v1/current-summoner", c.Port)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Basic Authentication başlığını ekliyoruz (username: riot, password: token)
	auth := fmt.Sprintf("riot:%s", c.Token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LCU hata kodu döndürdü: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var summoner SummonerInfo
	if err := json.Unmarshal(body, &summoner); err != nil {
		return nil, err
	}

	return &summoner, nil
}

func (c *Client) AcceptMatch() error {
	// Maç kabul etme endpoint'i
	url := fmt.Sprintf("https://127.0.0.1:%s/lol-matchmaking/v1/ready-check/accept", c.Port)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// Her zamanki gibi Basic Auth şifremizi ekliyoruz
	auth := fmt.Sprintf("riot:%s", c.Token)
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content veya 200 OK dönerse başarılıdır
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Println("✅ Maç başarıyla kabul edildi!")
		return nil
	}

	return fmt.Errorf("Maç kabul edilemedi. LCU hata kodu döndürdü: %d", resp.StatusCode)
}

func (c *Client) GetPickableChampions() ([]int, error) {
	url := fmt.Sprintf("https://127.0.0.1:%s/lol-champ-select/v1/pickable-champion-ids", c.Port)

	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.Token))
	req.Header.Add("Authorization", "Basic "+auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var champs []int
	json.Unmarshal(body, &champs)
	return champs, nil
}

func (c *Client) LockChampion(actionId int, championId int) error {
	url := fmt.Sprintf("https://127.0.0.1:%s/lol-champ-select/v1/session/actions/%d", c.Port, actionId)

	// completed: true yaparak şampiyonu hoverlamak yerine direkt kilitliyoruz
	payload := []byte(fmt.Sprintf(`{"championId": %d, "completed": true}`, championId))

	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.Token))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		fmt.Printf("✅ %d ID'li şampiyon başarıyla kilitlendi!\n", championId)
	}
	return nil
}
