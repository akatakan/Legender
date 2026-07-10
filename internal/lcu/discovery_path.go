package lcu

import (
	"os"
	"path/filepath"
)

func DefaultDiscoveryCachePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "lcu-lockfile-path"
	}
	return filepath.Join(configDir, "Legender", "lcu-lockfile-path")
}
