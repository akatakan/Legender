package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// İleride eklenecek tüm ayarları bu yapıya (struct) yazacağız
type Settings struct {
	AutoAccept bool             `json:"autoAccept"`
	AutoPick   AutoPickSettings `json:"autoPick"`
}

type AutoPickSettings struct {
	Enabled        bool  `json:"enabled"`
	PrimaryChamp   int   `json:"primaryChamp"`   // 1. Aşama (Örn: Yasuo = 157)
	SecondaryChamp int   `json:"secondaryChamp"` // 2. Aşama (Örn: Yone = 777)
	ChampPool      []int `json:"champPool"`      // 3. Aşama (Örn: [84, 238, 55])
}

// Tüm uygulamanın her yerinden erişilebilecek global ayar değişkenimiz
var AppSettings Settings

const settingsFile = "settings.json"

// Load, JSON dosyasını okur. Dosya yoksa oluşturur.
func Load() {
	file, err := os.ReadFile(settingsFile)
	if err != nil {
		fmt.Println("Ayarlar dosyası bulunamadı, yeni 'settings.json' oluşturuluyor...")
		// Dosya yoksa varsayılan değerleri ata ve kaydet
		AppSettings = Settings{
			AutoAccept: false,
			AutoPick: AutoPickSettings{
				Enabled:        false,
				PrimaryChamp:   0,
				SecondaryChamp: 0,
				ChampPool:      []int{},
			},
		}
		Save()
		return
	}

	// Dosya varsa içindeki veriyi AppSettings yapısına dönüştür
	err = json.Unmarshal(file, &AppSettings)
	if err != nil {
		fmt.Println("Ayarlar dosyası okunamadı:", err)
	}
}

// Save, mevcut AppSettings yapısını JSON formatında dosyaya yazar.
func Save() {
	// Veriyi şık (girintili) bir JSON formatına çeviriyoruz
	data, err := json.MarshalIndent(AppSettings, "", "  ")
	if err != nil {
		fmt.Println("Ayarlar kaydedilirken hata oluştu:", err)
		return
	}

	err = os.WriteFile(settingsFile, data, 0644)
	if err != nil {
		fmt.Println("Dosya yazma hatası:", err)
	}
}
