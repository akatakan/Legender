package main

import (
	"context"
	"fmt"
	"legender/internal/config"
	"legender/internal/diagnostics"
	"legender/internal/lcu"
	"log"
	"os"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	mu        sync.Mutex
	settings  *config.Store
	discovery *lcu.Discovery
	logFile   *os.File
	logPath   string
	lcuClient *lcu.Client
	lcuCancel context.CancelFunc
}

func NewApp() *App {
	return &App{
		settings:  config.NewStoreWithLegacy(config.DefaultPath(), "settings.json"),
		discovery: lcu.NewDiscovery(lcu.DefaultDiscoveryCachePath()),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logPath = diagnostics.DefaultLogPath()
	logFile, err := diagnostics.Setup(a.logPath)
	if err != nil {
		log.Println("Tanılama logu başlatılamadı:", err)
	} else {
		a.logFile = logFile
		log.Println("Legender başlatıldı")
	}
	if err := a.settings.Load(); err != nil {
		log.Println("Ayarlar yüklenemedi:", err)
	}
}

func (a *App) shutdown(context.Context) {
	a.stopLCUClient()
	if a.logFile != nil {
		log.Println("Legender kapatılıyor")
		_ = a.logFile.Close()
	}
}

type LCUResponse struct {
	IsActive    bool              `json:"isActive"`
	Port        string            `json:"port"`
	Locale      string            `json:"locale"`
	GameVersion string            `json:"gameVersion"`
	Summoner    *lcu.SummonerInfo `json:"summoner"`
}

func (a *App) GetLCUData() LCUResponse {
	info, err := a.discovery.GetConnectionInfo()
	if err != nil {
		log.Println("League Client bağlantı bilgisi alınamadı:", err)
		a.stopLCUClient()
		return LCUResponse{IsActive: false}
	}

	if !info.IsActive {
		a.stopLCUClient()
		return LCUResponse{IsActive: false}
	}

	a.mu.Lock()
	if a.lcuClient == nil || a.lcuClient.Port != info.Port || a.lcuClient.Token != info.Token {
		if a.lcuCancel != nil {
			a.lcuCancel()
		}
		a.lcuClient = lcu.NewClient(info.Port, info.Token)
		wsCtx, cancel := context.WithCancel(a.ctx)
		a.lcuCancel = cancel
		client := a.lcuClient
		go client.StartWebSocket(wsCtx, a.currentAutomationSettings)
	}
	client := a.lcuClient
	a.mu.Unlock()

	requestCtx, cancel := context.WithTimeout(a.ctx, 5*time.Second)
	defer cancel()
	summoner, err := client.GetCurrentSummoner(requestCtx)
	if err != nil {
		log.Println("Sihirdar bilgisi alınamadı:", err)
	}
	metadata, err := client.GetClientMetadata(requestCtx)
	if err != nil {
		log.Println("İstemci sürüm/locale bilgisi alınamadı:", err)
		metadata = &lcu.ClientMetadata{}
	}

	return LCUResponse{
		IsActive:    true,
		Port:        info.Port,
		Locale:      metadata.Locale,
		GameVersion: metadata.GameVersion,
		Summoner:    summoner,
	}
}

func (a *App) GetSettings() config.Settings {
	return a.settings.Snapshot()
}

func (a *App) ToggleAutoAccept(enabled bool) error {
	return a.settings.SetAutoAccept(enabled)
}

func (a *App) SaveAutoPickSettings(cfg config.AutoPickSettings) error {
	return a.settings.SetAutoPick(cfg)
}

func (a *App) ExportDiagnostics() (string, error) {
	if a.logPath == "" {
		return "", fmt.Errorf("tanılama logu henüz hazır değil")
	}
	destination, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           "Legender tanılama logunu kaydet",
		DefaultFilename: "legender-diagnostics.log",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Log dosyası (*.log)", Pattern: "*.log"},
		},
	})
	if err != nil || destination == "" {
		return destination, err
	}
	data, err := os.ReadFile(a.logPath)
	if err != nil {
		return "", fmt.Errorf("tanılama logu okunamadı: %w", err)
	}
	if err := os.WriteFile(destination, data, 0o600); err != nil {
		return "", fmt.Errorf("tanılama logu dışa aktarılamadı: %w", err)
	}
	return destination, nil
}

func (a *App) stopLCUClient() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.lcuCancel != nil {
		a.lcuCancel()
	}
	a.lcuCancel = nil
	a.lcuClient = nil
}

func (a *App) currentAutomationSettings() lcu.AutomationSettings {
	settings := a.settings.Snapshot()
	return lcu.AutomationSettings{
		AutoAccept: settings.AutoAccept,
		AutoPick: lcu.AutoPickSettings{
			Enabled:        settings.AutoPick.Enabled,
			PrimaryChamp:   settings.AutoPick.PrimaryChamp,
			SecondaryChamp: settings.AutoPick.SecondaryChamp,
			ChampPool:      settings.AutoPick.ChampPool,
		},
	}
}
