package httpapi

import "github.com/gofiber/fiber/v2"

// EventsApp serves ONLY the Hermes event callback. On Linux the container cannot reach an API bound to 127.0.0.1, so
// this narrow app is bound to the Docker bridge address instead (HERMES_EVENTS_LISTEN) and the full API stays on
// loopback. The route still requires the X-Hermes-Event-Key secret.
func (s *Server) EventsApp() *fiber.App {
	app := newFiber()
	app.Post("/api/v1/internal/hermes/events", s.hermesEvent)
	app.Post("/api/v1/internal/hermes/handoff", s.hermesHandoff)
	return app
}

// App builds the Fiber app with every route registered.
func (s *Server) App() *fiber.App {
	app := newFiber()

	// Public routes. They must be registered before the session middleware below.
	app.Get("/healthz", s.health)
	app.Post("/api/v1/auth/login", s.login)
	app.Post("/api/v1/auth/logout", s.logout)
	app.Get("/api/v1/public/p/:slug", s.getPublicLink) // Linktree-style page data; no session, only public fields
	app.Get("/api/v1/public/p/:slug/logo", s.getPublicLogo)
	app.Get("/api/v1/public/p/:slug/links/:kind", s.getPublicFolderLinks)
	app.Get("/api/v1/public/p", s.listPublicLinks)
	app.Post("/api/v1/internal/hermes/events", s.hermesEvent) // authenticated by X-Hermes-Event-Key, not a session
	app.Post("/api/v1/internal/hermes/handoff", s.hermesHandoff)

	api := app.Group("/api/v1")
	api.Use(s.authenticate)
	api.Get("/me", s.me)

	api.Get("/tenants", s.listTenants)
	api.Post("/tenants", s.createTenant)
	s.tenantRoutes(api.Group("/tenants/:id"))

	api.Get("/audit-logs", s.auditLogs)

	api.Get("/conversations", s.listConversations)
	api.Get("/conversations/stats", s.conversationStats)
	api.Patch("/conversations/:conversation_id", s.setConversationStatus)
	api.Get("/conversations/:conversation_id/messages", s.conversationMessages)

	api.Patch("/crawl/sources/:sid/force", s.requireSuperadmin, s.forceCrawlSource)

	hermesAdmin := api.Group("/hermes")
	hermesAdmin.Get("/overview", s.hermesOverview)
	hermesAdmin.Get("/skills", s.hermesSkills)
	hermesAdmin.Get("/connection", s.hermesConnection)
	hermesAdmin.Post("/sync-all", s.requireSuperadmin, s.startSyncAll)
	hermesAdmin.Get("/sync-all/:job_id", s.requireSuperadmin, s.syncAllStatus)
	hermesAdmin.Get("/settings", s.requireSuperadmin, s.getHermesSettings)
	hermesAdmin.Patch("/settings", s.requireSuperadmin, s.saveHermesSettings)
	return app
}

// tenantRoutes registers everything under /tenants/:id with the permission each one needs.
func (s *Server) tenantRoutes(t fiber.Router) {
	read := s.tenantAccess("tenant:read")
	member := s.tenantAccess("") // any member of the tenant
	allow := s.tenantAccess

	t.Get("/", read, s.getTenant)
	t.Patch("/", allow("tenant:update"), s.updateTenant)
	t.Post("/suspend", allow("tenant:lifecycle"), s.setTenantStatus("SUSPENDED"))
	t.Post("/resume", allow("tenant:lifecycle"), s.setTenantStatus("ONBOARDING"))

	t.Get("/members", allow("member:manage"), s.listMembers)
	t.Post("/members", allow("member:manage"), s.createMember)
	t.Patch("/members/:user_id", allow("member:manage"), s.updateMemberRole)

	t.Get("/profile", read, s.getProfile)
	t.Put("/profile", allow("tenant:update"), s.saveProfile)
	t.Put("/logo", allow("tenant:update"), s.uploadLogo)
	t.Delete("/logo", allow("tenant:update"), s.deleteLogo)

	t.Get("/bot", read, s.getBot)
	t.Put("/bot", allow("bot:update"), s.saveBot)
	t.Delete("/bot", allow("bot:update"), s.deleteBot)

	t.Get("/whatsapp", read, s.getWhatsApp)
	t.Put("/whatsapp", allow("bot:update"), s.saveWhatsApp)
	t.Put("/whatsapp/service", allow("bot:update"), s.setWhatsAppService)
	t.Get("/whatsapp/admins", read, s.listWhatsAppAdmins)
	t.Put("/whatsapp/admins", allow("bot:update"), s.saveWhatsAppAdmins)
	t.Delete("/whatsapp", allow("bot:update"), s.deleteWhatsApp)
	t.Post("/whatsapp/pairing", allow("bot:update"), s.startWhatsAppPairing)
	t.Get("/whatsapp/pairing", allow("bot:update"), s.whatsappPairingStatus)
	t.Delete("/whatsapp/pairing", allow("bot:update"), s.cancelWhatsAppPairing)

	t.Get("/files", read, s.listFiles)
	t.Post("/files", allow("file:create"), s.uploadFile)
	t.Get("/files/:file_id", read, s.downloadFile)
	t.Get("/files/:file_id/download", read, s.downloadFile)
	t.Patch("/files/:file_id", allow("file:update"), s.renameFile)
	t.Delete("/files/:file_id", allow("file:delete"), s.deleteFile)
	t.Get("/folders", read, s.listFolders)
	t.Post("/folders", allow("file:create"), s.createFolder)

	t.Get("/links", read, s.listLinks)
	t.Post("/links", allow("tenant:update"), s.createLink)
	t.Patch("/links/:lid", allow("tenant:update"), s.updateLink)
	t.Delete("/links/:lid", allow("tenant:update"), s.deleteLink)
	t.Get("/questions", read, s.listQuestions)
	t.Patch("/questions/:qid", allow("tenant:update"), s.updateQuestion)
	t.Delete("/questions/:qid", allow("tenant:update"), s.deleteQuestion)
	t.Get("/contacts", read, s.listContacts)
	t.Post("/contacts", allow("tenant:update"), s.createContact)
	t.Patch("/contacts/:cid", allow("tenant:update"), s.updateContact)
	t.Delete("/contacts/:cid", allow("tenant:update"), s.deleteContact)

	t.Get("/crawl/sources", read, s.listCrawlSources)
	t.Post("/crawl/sources", allow("tenant:update"), s.createCrawlSource)
	t.Patch("/crawl/sources/:sid", allow("tenant:update"), s.updateCrawlSource)
	t.Delete("/crawl/sources/:sid", allow("tenant:update"), s.deleteCrawlSource)
	t.Post("/crawl/sources/:sid/run", allow("tenant:update"), s.runCrawlSource)
	t.Get("/crawl/sources/:sid/runs", read, s.listCrawlRuns)
	t.Get("/crawl/sources/:sid/events", read, s.listCrawlEvents)
	t.Get("/crawl/sources/:sid/pages", read, s.listCrawlPages)
	t.Get("/crawl/events/:eid", read, s.getCrawlEvent)
	t.Get("/crawl/pages/:pid", read, s.getCrawlPage)
	t.Put("/crawl/pages/:pid/content", allow("tenant:update"), s.saveCrawlPageContent)
	t.Post("/crawl/pages/:pid/resolve", allow("tenant:update"), s.resolveCrawlPage)
	t.Patch("/crawl/pages/:pid", allow("tenant:update"), s.patchCrawlPage)

	t.Post("/hermes/provision", allow("assistant:manage"), s.provisionAssistant)
	t.Post("/hermes/action", allow("assistant:manage"), s.assistantAction)
	t.Post("/hermes/sync", allow("file:update"), s.syncKnowledge)
	t.Get("/knowledge/status", read, s.knowledgeStatus)

	t.Post("/telegram/dev-message", allow("bot:update"), s.devTelegramMessage)
	t.Get("/telegram/export", member, s.exportTelegramLog)
	t.Get("/telegram/conversations", member, s.telegramConversationGroups)

	t.All("/*", read, notImplemented)
}
