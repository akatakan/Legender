package main

import (
	"context"
	"legender/internal/config"
	"legender/internal/lcu"
)

type App struct {
	ctx       context.Context
	lcuClient *lcu.Client
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	config.Load()
}

type LCUResponse struct {
	IsActive bool              `json:"isActive"`
	Port     string            `json:"port"`
	Summoner *lcu.SummonerInfo `json:"summoner"`
}

func (a *App) GetLCUData() LCUResponse {
	info := lcu.GetConnectionInfo()

	if !info.IsActive {
		a.lcuClient = nil // Oyun kapandıysa istemciyi sıfırla
		return LCUResponse{IsActive: false}
	}

	// Eğer yeni açıldıysa veya henüz dinlemeye başlamadıysak başlat
	if a.lcuClient == nil || a.lcuClient.Port != info.Port {
		a.lcuClient = lcu.NewClient(info.Port, info.Token)

		// WebSocket'i arka planda kitlemeden (goroutine) başlatıyoruz
		go a.lcuClient.StartWebSocket()
	}

	summoner, _ := a.lcuClient.GetCurrentSummoner()

	return LCUResponse{
		IsActive: true,
		Port:     info.Port,
		Summoner: summoner,
	}
}

func (a *App) GetSettings() config.Settings {
	return config.AppSettings
}

func (a *App) ToggleAutoAccept(enabled bool) {
	config.AppSettings.AutoAccept = enabled
	config.Save()
}

func (a *App) SaveAutoPickSettings(cfg config.AutoPickSettings) {
	config.AppSettings.AutoPick = cfg
	config.Save()
}
