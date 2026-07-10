package diagnostics

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupRotatesOversizedLog(t *testing.T) {
	oldWriter := log.Writer()
	oldFlags := log.Flags()
	defer func() {
		log.SetOutput(oldWriter)
		log.SetFlags(oldFlags)
	}()

	path := filepath.Join(t.TempDir(), "logs", "legender.log")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, maxLogSize), 0o600); err != nil {
		t.Fatal(err)
	}

	file, err := Setup(path)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	log.Print("new log")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("rotated log not found: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("new log file is empty")
	}
	log.SetOutput(io.Discard)
}
