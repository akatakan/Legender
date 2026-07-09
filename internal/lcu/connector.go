package lcu

import (
	"os/exec"
	"regexp"
	"strings"
)

type Info struct {
	Port     string `json:"port"`
	Token    string `json:"token"`
	IsActive bool   `json:"isActive"`
}

func GetConnectionInfo() Info {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-CimInstance Win32_Process -Filter \"Name='LeagueClientUx.exe'\" | Select-Object -ExpandProperty CommandLine")
	out, err := cmd.Output()

	if err != nil {
		return Info{IsActive: false}
	}

	output := string(out)

	if !strings.Contains(output, "--app-port") {
		return Info{IsActive: false}
	}

	portRegex := regexp.MustCompile(`--app-port=([0-9]+)`)
	portMatch := portRegex.FindStringSubmatch(output)

	tokenRegex := regexp.MustCompile(`--remoting-auth-token=([a-zA-Z0-9_-]+)`)
	tokenMatch := tokenRegex.FindStringSubmatch(output)

	if len(portMatch) > 1 && len(tokenMatch) > 1 {
		return Info{
			Port:     portMatch[1],
			Token:    tokenMatch[1],
			IsActive: true,
		}
	}

	return Info{IsActive: false}
}
