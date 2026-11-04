package httpapi

import (
	"net/http/httptest"
	"testing"

	"ponpes-saas/internal/config"
	"ponpes-saas/internal/repository"
	"ponpes-saas/internal/storage"
)

func TestRoutesAreRegistered(t *testing.T) {
	app := NewServer(config.Config{}, repository.Repository{}, storage.New(t.TempDir())).App()
	registered := map[string]bool{}
	for _, route := range app.GetRoutes() {
		registered[route.Method+" "+route.Path] = true
	}
	want := []string{
		"GET /healthz", "POST /api/v1/auth/login", "POST /api/v1/auth/logout", "POST /api/v1/internal/hermes/events",
		"GET /api/v1/me", "GET /api/v1/tenants", "POST /api/v1/tenants",
		"GET /api/v1/tenants/:id/", "PATCH /api/v1/tenants/:id/", "POST /api/v1/tenants/:id/suspend", "POST /api/v1/tenants/:id/resume",
		"GET /api/v1/tenants/:id/members", "POST /api/v1/tenants/:id/members", "PATCH /api/v1/tenants/:id/members/:user_id",
		"GET /api/v1/tenants/:id/profile", "PUT /api/v1/tenants/:id/profile",
		"GET /api/v1/tenants/:id/bot", "PUT /api/v1/tenants/:id/bot", "DELETE /api/v1/tenants/:id/bot",
		"GET /api/v1/tenants/:id/files", "POST /api/v1/tenants/:id/files", "GET /api/v1/tenants/:id/files/:file_id",
		"GET /api/v1/tenants/:id/files/:file_id/download", "PATCH /api/v1/tenants/:id/files/:file_id", "DELETE /api/v1/tenants/:id/files/:file_id",
		"GET /api/v1/tenants/:id/folders", "POST /api/v1/tenants/:id/folders",
		"POST /api/v1/tenants/:id/hermes/provision", "POST /api/v1/tenants/:id/hermes/action", "POST /api/v1/tenants/:id/hermes/sync",
		"GET /api/v1/tenants/:id/knowledge/status", "POST /api/v1/tenants/:id/telegram/dev-message",
		"GET /api/v1/tenants/:id/telegram/export", "GET /api/v1/tenants/:id/telegram/conversations",
		"GET /api/v1/audit-logs", "GET /api/v1/conversations", "GET /api/v1/conversations/stats",
		"PATCH /api/v1/conversations/:conversation_id", "GET /api/v1/conversations/:conversation_id/messages",
		"GET /api/v1/hermes/overview", "GET /api/v1/hermes/skills", "GET /api/v1/hermes/connection",
		"POST /api/v1/hermes/sync-all", "GET /api/v1/hermes/sync-all/:job_id", "GET /api/v1/hermes/settings", "PATCH /api/v1/hermes/settings",
		"GET /api/v1/tenants/:id/crawl/sources", "POST /api/v1/tenants/:id/crawl/sources", "PATCH /api/v1/tenants/:id/crawl/sources/:sid",
		"DELETE /api/v1/tenants/:id/crawl/sources/:sid", "POST /api/v1/tenants/:id/crawl/sources/:sid/run", "PATCH /api/v1/crawl/sources/:sid/force",
		"GET /api/v1/tenants/:id/crawl/sources/:sid/runs", "GET /api/v1/tenants/:id/crawl/sources/:sid/events", "GET /api/v1/tenants/:id/crawl/sources/:sid/pages",
		"GET /api/v1/tenants/:id/crawl/events/:eid", "GET /api/v1/tenants/:id/crawl/pages/:pid", "PUT /api/v1/tenants/:id/crawl/pages/:pid/content",
		"POST /api/v1/tenants/:id/crawl/pages/:pid/resolve", "PATCH /api/v1/tenants/:id/crawl/pages/:pid",
		"GET /api/v1/tenants/:id/links", "POST /api/v1/tenants/:id/links", "PATCH /api/v1/tenants/:id/links/:lid", "DELETE /api/v1/tenants/:id/links/:lid",
	}
	for _, route := range want {
		if !registered[route] {
			t.Errorf("route not registered: %s", route)
		}
	}
}

func TestHealthzNeedsNoSession(t *testing.T) {
	app := NewServer(config.Config{}, repository.Repository{}, storage.New(t.TempDir())).App()
	response, err := app.Test(httptest.NewRequest("GET", "/healthz", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 {
		t.Fatalf("healthz status = %d, want 200", response.StatusCode)
	}
}
