//go:build windows

package lcu

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

const stillActiveExitCode = 259

type Info struct {
	Port     string `json:"port"`
	Token    string `json:"token"`
	IsActive bool   `json:"isActive"`
}

type Discovery struct {
	mu                   sync.Mutex
	cachePath            string
	lockfilePath         string
	lastDiscoveryAttempt time.Time
}

type leagueProcess struct {
	ProcessID      int    `json:"ProcessId"`
	ExecutablePath string `json:"ExecutablePath"`
	CommandLine    string `json:"CommandLine"`
}

func NewDiscovery(cachePath string) *Discovery {
	discovery := &Discovery{cachePath: cachePath}
	if data, err := os.ReadFile(cachePath); err == nil {
		discovery.lockfilePath = strings.TrimSpace(string(data))
	}
	return discovery
}

func (d *Discovery) GetConnectionInfo() (Info, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lockfilePath != "" {
		info, err := connectionFromLockfile(d.lockfilePath)
		if err == nil && info.IsActive {
			return info, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return Info{IsActive: false}, err
		}
		if time.Since(d.lastDiscoveryAttempt) < 30*time.Second {
			return Info{IsActive: false}, nil
		}
	}
	if !d.lastDiscoveryAttempt.IsZero() && time.Since(d.lastDiscoveryAttempt) < 10*time.Second {
		return Info{IsActive: false}, nil
	}

	d.lastDiscoveryAttempt = time.Now()
	process, active, err := discoverLeagueProcess()
	if err != nil || !active {
		return Info{IsActive: false}, err
	}

	if process.ExecutablePath != "" {
		d.rememberLockfile(filepath.Join(filepath.Dir(process.ExecutablePath), "lockfile"))
		info, lockfileErr := connectionFromLockfile(d.lockfilePath)
		if lockfileErr == nil {
			return info, nil
		}
		if !errors.Is(lockfileErr, os.ErrNotExist) {
			return Info{IsActive: false}, lockfileErr
		}
	}

	return connectionFromCommandLine(process.CommandLine)
}

func (d *Discovery) rememberLockfile(path string) {
	d.lockfilePath = path
	if d.cachePath == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(d.cachePath), 0o700); err != nil {
		return
	}
	_ = os.WriteFile(d.cachePath, []byte(path), 0o600)
}

func discoverLeagueProcess() (leagueProcess, bool, error) {
	command := `Get-CimInstance Win32_Process -Filter "Name='LeagueClientUx.exe'" | Select-Object -First 1 ProcessId,ExecutablePath,CommandLine | ConvertTo-Json -Compress`
	out, err := exec.Command("powershell", "-NoProfile", "-Command", command).Output()
	if err != nil {
		return leagueProcess{}, false, fmt.Errorf("LeagueClientUx.exe sorgulanamadı: %w", err)
	}
	if strings.TrimSpace(string(out)) == "" {
		return leagueProcess{}, false, nil
	}
	var process leagueProcess
	if err := json.Unmarshal(out, &process); err != nil {
		return leagueProcess{}, false, fmt.Errorf("LeagueClientUx.exe bilgisi çözülemedi: %w", err)
	}
	return process, true, nil
}

func connectionFromLockfile(path string) (Info, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Info{IsActive: false}, err
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), ":", 5)
	if len(parts) != 5 {
		return Info{IsActive: false}, fmt.Errorf("League lockfile formatı geçersiz")
	}
	pid, err := strconv.Atoi(parts[1])
	if err != nil || pid <= 0 {
		return Info{IsActive: false}, fmt.Errorf("League lockfile PID değeri geçersiz")
	}
	if !processRunning(uint32(pid)) {
		return Info{IsActive: false}, nil
	}
	if parts[2] == "" || parts[3] == "" {
		return Info{IsActive: false}, fmt.Errorf("League lockfile port veya token içermiyor")
	}
	return Info{Port: parts[2], Token: parts[3], IsActive: true}, nil
}

func connectionFromCommandLine(commandLine string) (Info, error) {
	if !strings.Contains(commandLine, "--app-port") {
		return Info{IsActive: false}, nil
	}
	portMatch := regexp.MustCompile(`--app-port=([0-9]+)`).FindStringSubmatch(commandLine)
	tokenMatch := regexp.MustCompile(`--remoting-auth-token=([a-zA-Z0-9_-]+)`).FindStringSubmatch(commandLine)
	if len(portMatch) <= 1 || len(tokenMatch) <= 1 {
		return Info{IsActive: false}, fmt.Errorf("League Client port veya token bilgisi ayrıştırılamadı")
	}
	return Info{Port: portMatch[1], Token: tokenMatch[1], IsActive: true}, nil
}

func processRunning(pid uint32) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)
	var exitCode uint32
	if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}
	return exitCode == stillActiveExitCode
}
