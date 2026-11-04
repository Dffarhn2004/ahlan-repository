package httpapi

import (
	"net/http/httptest"
	"testing"

	"ponpes-saas/internal/config"
	"ponpes-saas/internal/repository"
	"ponpes-saas/internal/storage"
)

// The listener exposed to the Hermes container must serve the event callback and nothing else.
func TestEventsAppExposesOnlyTheEventRoute(t *testing.T) {
	app := NewServer(config.Config{HermesAPIKey: "k-for-test-123"}, repository.Repository{}, storage.New(t.TempDir())).EventsApp()
	routes := 0
	for _, route := range app.GetRoutes(true) {
		if route.Path != "/" || route.Method != "HEAD" {
			routes++
			if route.Path != "/api/v1/internal/hermes/events" {
				t.Errorf("unexpected route on the events listener: %s %s", route.Method, route.Path)
			}
		}
	}
	for _, path := range []string{"/healthz", "/api/v1/tenants", "/api/v1/auth/login", "/api/v1/hermes/settings"} {
		resp, err := app.Test(httptest.NewRequest("GET", path, nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 404 && resp.StatusCode != 405 {
			t.Errorf("%s should not be served on the events listener, got %d", path, resp.StatusCode)
		}
	}
	// Without the shared key the callback is refused.
	resp, err := app.Test(httptest.NewRequest("POST", "/api/v1/internal/hermes/events", nil))
	if err != nil || resp.StatusCode == 200 {
		t.Fatalf("callback without key must be refused, got %v %v", resp, err)
	}
}
