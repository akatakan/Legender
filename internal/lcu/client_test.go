package lcu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestGetClientMetadataUsesLocaleVersionAndCache(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/riotclient/region-locale":
			_, _ = w.Write([]byte(`{"locale":"tr_TR","region":"TR1"}`))
		case "/lol-patch/v1/game-version":
			_, _ = w.Write([]byte(`"16.13.123.456"`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := testClientForServer(t, server)
	for i := 0; i < 2; i++ {
		metadata, err := client.GetClientMetadata(context.Background())
		if err != nil {
			t.Fatalf("GetClientMetadata() error = %v", err)
		}
		if metadata.Locale != "tr_TR" || metadata.Region != "TR1" || metadata.GameVersion != "16.13.123.456" {
			t.Fatalf("metadata = %+v", metadata)
		}
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("metadata endpoint requests = %d, want 2 total", got)
	}
}
