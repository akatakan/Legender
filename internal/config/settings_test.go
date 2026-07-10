package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestStoreCreatesDefaultsAndPersistsUpdates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := NewStore(path)

	if err := store.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if err := store.SetAutoAccept(true); err != nil {
		t.Fatalf("SetAutoAccept() error = %v", err)
	}
	if err := store.SetAutoPick(AutoPickSettings{Enabled: true, PrimaryChamp: 157, ChampPool: []int{84, 238}}); err != nil {
		t.Fatalf("SetAutoPick() error = %v", err)
	}

	reloaded := NewStore(path)
	if err := reloaded.Load(); err != nil {
		t.Fatalf("reloaded Load() error = %v", err)
	}
	got := reloaded.Snapshot()
	if !got.AutoAccept || !got.AutoPick.Enabled || got.AutoPick.PrimaryChamp != 157 || len(got.AutoPick.ChampPool) != 2 {
		t.Fatalf("persisted settings = %+v", got)
	}
}

func TestSnapshotDoesNotExposeInternalSlice(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "settings.json"))
	if err := store.SetAutoPick(AutoPickSettings{ChampPool: []int{1, 2}}); err != nil {
		t.Fatal(err)
	}

	snapshot := store.Snapshot()
	snapshot.AutoPick.ChampPool[0] = 999
	if got := store.Snapshot().AutoPick.ChampPool[0]; got != 1 {
		t.Fatalf("internal slice changed through snapshot: %d", got)
	}
}

func TestConcurrentUpdatesLeaveValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := NewStore(path)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(enabled bool) {
			defer wg.Done()
			if err := store.SetAutoAccept(enabled); err != nil {
				t.Errorf("SetAutoAccept() error = %v", err)
			}
		}(i%2 == 0)
		go func(champion int) {
			defer wg.Done()
			if err := store.SetAutoPick(AutoPickSettings{PrimaryChamp: champion, ChampPool: []int{champion}}); err != nil {
				t.Errorf("SetAutoPick() error = %v", err)
			}
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var persisted Settings
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatalf("persisted JSON is invalid: %v", err)
	}
}

func TestStoreMigratesLegacySettings(t *testing.T) {
	root := t.TempDir()
	legacyPath := filepath.Join(root, "legacy.json")
	newPath := filepath.Join(root, "nested", "Legender", "settings.json")
	legacy := []byte(`{"autoAccept":true,"autoPick":{"enabled":true,"primaryChamp":99,"secondaryChamp":0,"champPool":[1,2]}}`)
	if err := os.WriteFile(legacyPath, legacy, 0o600); err != nil {
		t.Fatal(err)
	}

	store := NewStoreWithLegacy(newPath, legacyPath)
	if err := store.Load(); err != nil {
		t.Fatalf("Load() migration error = %v", err)
	}
	if got := store.Snapshot(); !got.AutoAccept || got.AutoPick.PrimaryChamp != 99 {
		t.Fatalf("migrated settings = %+v", got)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("new settings file was not created: %v", err)
	}
	if _, err := os.Stat(legacyPath); err != nil {
		t.Fatalf("legacy settings file should be preserved: %v", err)
	}
}

func TestStoreRejectsNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"schemaVersion":999}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := NewStore(path).Load(); err == nil {
		t.Fatal("Load() accepted an unsupported schema version")
	}
}

func TestLegacySchemaNormalizesChampionPool(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	data := []byte(`{"autoPick":{"primaryChamp":-1,"secondaryChamp":2,"champPool":[0,5,5,-3,6]}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewStore(path)
	if err := store.Load(); err != nil {
		t.Fatal(err)
	}
	got := store.Snapshot()
	if got.SchemaVersion != currentSchemaVersion || got.AutoPick.PrimaryChamp != 0 {
		t.Fatalf("migrated settings = %+v", got)
	}
	if len(got.AutoPick.ChampPool) != 2 || got.AutoPick.ChampPool[0] != 5 || got.AutoPick.ChampPool[1] != 6 {
		t.Fatalf("normalized pool = %v", got.AutoPick.ChampPool)
	}
}
