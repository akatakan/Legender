package lcu

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"legender/internal/config"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
)

// WAMP protokolü için gelen mesaj yapısı
type WAMPMessage []interface{}

// LCU'dan gelen Ready Check (Maç Bulundu) verisi
type ReadyCheckData struct {
	State string `json:"state"`
}

// StartWebSocket, LCU'ya bağlanır ve eventleri dinlemeye başlar
func (c *Client) StartWebSocket() {
	url := fmt.Sprintf("wss://127.0.0.1:%s/", c.Port)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	auth := base64.StdEncoding.EncodeToString([]byte("riot:" + c.Token))
	headers := http.Header{"Authorization": []string{"Basic " + auth}}

	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		log.Println("WebSocket bağlantı hatası:", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ WebSocket Bağlantısı Kuruldu. Eventler dinleniyor...")

	// 1. Sadece Maç Bulunma (ready-check) eventine abone oluyoruz
	subscribeMsg := `[5, "OnJsonApiEvent_lol-matchmaking_v1_ready-check"]`
	conn.WriteMessage(websocket.TextMessage, []byte(subscribeMsg))

	// 2. Sonsuz döngüde gelen mesajları dinliyoruz
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket koptu:", err)
			break
		}

		// Gelen mesajı WAMP formatına (JSON array) çevir
		var msg WAMPMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// WAMP Event mesajları 3 elemanlıdır: [8, "EventAdı", {Data}]
		if len(msg) == 3 {
			eventName, ok := msg[1].(string)
			if ok && eventName == "OnJsonApiEvent_lol-matchmaking_v1_ready-check" {

				// 1. Gelen payload bir obje (map) mi diye GÜVENLİ bir şekilde kontrol et
				eventPayload, ok := msg[2].(map[string]interface{})
				if !ok || eventPayload["data"] == nil {
					continue // Data yoksa es geç, çökmesini engelle
				}

				// 2. Data'nın içini GÜVENLİ bir şekilde aç
				dataField, ok := eventPayload["data"].(map[string]interface{})
				if !ok {
					continue
				}

				// 3. State ve PlayerResponse alanlarını güvenle al
				state, _ := dataField["state"].(string)
				playerResponse, _ := dataField["playerResponse"].(string)

				// 4. Eğer durum InProgress ise, Toggle Açıksa ve BİZ HENÜZ KABUL ETMEDİYSEK:
				if state == "InProgress" && config.AppSettings.AutoAccept && playerResponse == "None" {
					fmt.Println("🚀 Maç bulundu! Otomatik kabul ediliyor...")
					c.AcceptMatch()
				}
			}

			if ok && eventName == "OnJsonApiEvent_lol-champ-select_v1_session" {

				eventPayload, ok := msg[2].(map[string]interface{})
				if !ok || eventPayload["data"] == nil {
					continue
				}

				// Veriyi map'ler içinde kaybolmadan güvenle okumak için JSON'a çevirip kendi yapımıza aktarıyoruz
				jsonData, _ := json.Marshal(eventPayload["data"])

				var session struct {
					Timer struct {
						Phase string `json:"phase"`
					} `json:"timer"`
					LocalPlayerCellId int `json:"localPlayerCellId"`
					Actions           [][]struct {
						Id           int    `json:"id"`
						ActorCellId  int    `json:"actorCellId"`
						Type         string `json:"type"`
						Completed    bool   `json:"completed"`
						IsInProgress bool   `json:"isInProgress"`
					} `json:"actions"`
				}
				json.Unmarshal(jsonData, &session)

				// Eğer AutoPick kapalıysa veya seçim aşamasında değilsek hiçbir şey yapma
				cfg := config.AppSettings.AutoPick
				if !cfg.Enabled || session.Timer.Phase != "PLANNING" {
					continue
				}

				// Sıranın bizde olup olmadığını buluyoruz
				for _, actionGroup := range session.Actions {
					for _, action := range actionGroup {
						// Eğer sıra bizdeyse, işlem "pick" (seçim) ise ve henüz kilitlemediysek:
						if action.ActorCellId == session.LocalPlayerCellId && action.Type == "pick" && action.IsInProgress && !action.Completed {

							fmt.Println("⏳ Sıra bizde! 3 Aşamalı Seçim Algoritması başlıyor...")

							// 1. O an seçilebilir olan şampiyonları API'den çek
							pickableChamps, err := c.GetPickableChampions()
							if err != nil {
								continue
							}

							// Helper: Şampiyonun seçilebilir olup olmadığını kontrol eden küçük fonksiyon
							isPickable := func(champId int) bool {
								for _, id := range pickableChamps {
									if id == champId {
										return true
									}
								}
								return false
							}

							var targetChamp int

							// AŞAMA 1: Birincil Şampiyon
							if isPickable(cfg.PrimaryChamp) {
								targetChamp = cfg.PrimaryChamp
								fmt.Println("-> Aşama 1: Birincil şampiyon seçilebilir durumda.")
							} else if isPickable(cfg.SecondaryChamp) {
								// AŞAMA 2: İkincil Şampiyon
								targetChamp = cfg.SecondaryChamp
								fmt.Println("-> Aşama 2: Birincil kapalı, İkincil şampiyon seçiliyor.")
							} else if len(cfg.ChampPool) > 0 {
								// AŞAMA 3: Rastgele Havuz
								var validPool []int
								for _, id := range cfg.ChampPool {
									if isPickable(id) {
										validPool = append(validPool, id)
									}
								}
								if len(validPool) > 0 {
									targetChamp = validPool[rand.Intn(len(validPool))]
									fmt.Println("-> Aşama 3: Havuzdan rastgele bir şampiyon seçiliyor.")
								}
							}

							// Eğer hedef şampiyon belirlendiyse, kilitle!
							if targetChamp != 0 {
								c.LockChampion(action.Id, targetChamp)
							}
						}
					}
				}
			}
		}
	}
}
