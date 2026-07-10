package diagnostics

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const maxLogSize = 2 * 1024 * 1024

func DefaultLogPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "legender.log"
	}
	return filepath.Join(configDir, "Legender", "logs", "legender.log")
}

func Setup(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("log dizini oluşturulamadı: %w", err)
	}
	if info, err := os.Stat(path); err == nil && info.Size() >= maxLogSize {
		rotated := path + ".1"
		_ = os.Remove(rotated)
		if err := os.Rename(path, rotated); err != nil {
			return nil, fmt.Errorf("log döndürülemedi: %w", err)
		}
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("log dosyası açılamadı: %w", err)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(io.MultiWriter(os.Stderr, file))
	return file, nil
}
