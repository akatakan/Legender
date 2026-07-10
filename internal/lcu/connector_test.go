//go:build windows

package lcu

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConnectionFromCommandLine(t *testing.T) {
	info, err := connectionFromCommandLine(`LeagueClientUx.exe --app-port=54321 --remoting-auth-token=test_token-1`)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsActive || info.Port != "54321" || info.Token != "test_token-1" {
		t.Fatalf("connection info = %+v", info)
	}
}

func TestConnectionFromCommandLineRejectsIncompleteData(t *testing.T) {
	if _, err := connectionFromCommandLine(`LeagueClientUx.exe --app-port=54321`); err == nil {
		t.Fatal("connectionFromCommandLine() accepted missing token")
	}
}

func TestDiscoveryPersistsAndReloadsLockfilePath(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "cache", "lcu-lockfile-path")
	lockfilePath := `C:\Riot Games\League of Legends\lockfile`
	discovery := NewDiscovery(cachePath)
	discovery.rememberLockfile(lockfilePath)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != lockfilePath {
		t.Fatalf("cached path = %q", data)
	}
	if got := NewDiscovery(cachePath).lockfilePath; got != lockfilePath {
		t.Fatalf("reloaded path = %q, want %q", got, lockfilePath)
	}
}
