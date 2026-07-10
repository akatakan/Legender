package lcu

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestChooseChampionPriority(t *testing.T) {
	tests := []struct {
		name     string
		cfg      AutoPickSettings
		pickable []int
		want     int
	}{
		{name: "primary", cfg: AutoPickSettings{PrimaryChamp: 10, SecondaryChamp: 20, ChampPool: []int{30}}, pickable: []int{10, 20, 30}, want: 10},
		{name: "secondary", cfg: AutoPickSettings{PrimaryChamp: 10, SecondaryChamp: 20, ChampPool: []int{30}}, pickable: []int{20, 30}, want: 20},
		{name: "pool", cfg: AutoPickSettings{PrimaryChamp: 10, SecondaryChamp: 20, ChampPool: []int{30, 40}}, pickable: []int{40}, want: 40},
		{name: "none", cfg: AutoPickSettings{PrimaryChamp: 10, SecondaryChamp: 20, ChampPool: []int{30}}, pickable: []int{99}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chooseChampion(tt.cfg, tt.pickable, func(int) int { return 0 }); got != tt.want {
				t.Fatalf("chooseChampion() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestChampSelectEventLocksChampionDuringBanPick(t *testing.T) {
	var lockedChampion int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/lol-champ-select/v1/pickable-champion-ids":
			_, _ = w.Write([]byte(`[157]`))
		case r.Method == http.MethodPatch && r.URL.Path == "/lol-champ-select/v1/session/actions/42":
			var payload struct {
				ChampionID int `json:"championId"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("PATCH payload decode error = %v", err)
			}
			lockedChampion = payload.ChampionID
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := testClientForServer(t, server)
	message := []byte(fmt.Sprintf(`[8,%q,{"eventType":"Update","data":{"gameId":7,"timer":{"phase":"BAN_PICK"},"localPlayerCellId":3,"actions":[[{"id":42,"actorCellId":3,"type":"pick","completed":false,"isInProgress":true}]]}}]`, champSelectEvent))
	state := &automationState{handledActions: make(map[string]struct{})}
	settings := AutomationSettings{AutoPick: AutoPickSettings{Enabled: true, PrimaryChamp: 157}}

	if err := client.handleWAMPMessage(context.Background(), message, settings, state); err != nil {
		t.Fatalf("handleWAMPMessage() error = %v", err)
	}
	if lockedChampion != 157 {
		t.Fatalf("locked champion = %d, want 157", lockedChampion)
	}

	lockedChampion = 0
	if err := client.handleWAMPMessage(context.Background(), message, settings, state); err != nil {
		t.Fatalf("duplicate handleWAMPMessage() error = %v", err)
	}
	if lockedChampion != 0 {
		t.Fatalf("duplicate event locked champion again: %d", lockedChampion)
	}
}

func testClientForServer(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		t.Fatal(err)
	}
	return &Client{Port: port, Token: "test-token", httpClient: server.Client()}
}
