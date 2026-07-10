//go:build windows

package lcu

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Bu test yalnızca açık gerçek istemcide salt-okunur GET çağrıları yapar.
// Kuyruk, ready-check veya champ-select endpointlerine istek göndermez.
func TestRealLCUReadOnly(t *testing.T) {
	if os.Getenv("LEGENDER_LCU_INTEGRATION") != "1" {
		t.Skip("set LEGENDER_LCU_INTEGRATION=1 to test an open League Client")
	}

	discovery := NewDiscovery(filepath.Join(t.TempDir(), "lcu-lockfile-path"))
	info, err := discovery.GetConnectionInfo()
	if err != nil {
		t.Fatalf("GetConnectionInfo() error = %v", err)
	}
	if !info.IsActive {
		t.Fatal("League Client is not active")
	}

	client := NewClient(info.Port, info.Token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	metadata, err := client.GetClientMetadata(ctx)
	if err != nil {
		t.Fatalf("GetClientMetadata() error = %v", err)
	}
	if metadata.Locale == "" || metadata.Region == "" || metadata.GameVersion == "" {
		t.Fatalf("client metadata is incomplete: locale=%q region=%q versionPresent=%t", metadata.Locale, metadata.Region, metadata.GameVersion != "")
	}

	summoner, err := client.GetCurrentSummoner(ctx)
	if err != nil {
		t.Fatalf("GetCurrentSummoner() error = %v", err)
	}
	if summoner == nil || summoner.ProfileIconId < 0 {
		t.Fatal("summoner response is incomplete")
	}
}
