//go:build !windows

package lcu

import (
	"fmt"
	"runtime"
)

type Info struct {
	Port     string `json:"port"`
	Token    string `json:"token"`
	IsActive bool   `json:"isActive"`
}

type Discovery struct{}

func NewDiscovery(string) *Discovery {
	return &Discovery{}
}

func (d *Discovery) GetConnectionInfo() (Info, error) {
	return Info{IsActive: false}, fmt.Errorf("League Client keşfi %s platformunda desteklenmiyor", runtime.GOOS)
}
