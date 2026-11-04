package hermes

// Wire format of the Hermes bridge. Every operation is a fixed route with a typed request; there is deliberately no
// route that accepts a command, a container name, a Docker argument or a free-form path.

const (
	bridgeProfileCreate    = "/v1/profile/create"
	bridgeProfileAction    = "/v1/profile/action"
	bridgeProfileSettings  = "/v1/profile/settings"
	bridgeClientExperience = "/v1/profile/client-experience"
	bridgeWelcomeHook      = "/v1/profile/welcome-hook"
	bridgeLogHook          = "/v1/profile/log-hook"
	bridgeTelegram         = "/v1/channel/telegram"
	bridgeWhatsApp         = "/v1/channel/whatsapp"
	bridgeChannelRemove    = "/v1/channel/remove"
	bridgePair             = "/v1/whatsapp/pair"
	bridgePairStop         = "/v1/whatsapp/pair/stop"
	bridgeSessionRemove    = "/v1/whatsapp/session/remove"
	bridgeKnowledgePut     = "/v1/knowledge/put"
	bridgeKnowledgePrune   = "/v1/knowledge/prune"
	bridgeKnowledgeExtract = "/v1/knowledge/extract"
	bridgeKnowledgeReady   = "/v1/knowledge/ready"
	bridgeHealth           = "/v1/health"
	// bridgeGatewayRestart restarts the one shared gateway process. It takes no arguments on purpose.
	bridgeGatewayRestart = "/v1/gateway/restart"
)

type reqProfile struct {
	ProfileKey string `json:"profile_key"`
}

type reqCreate struct {
	TenantID   string `json:"tenant_id"`
	ProfileKey string `json:"profile_key"`
}

type reqAction struct {
	ProfileKey string `json:"profile_key"`
	Action     string `json:"action"`
}

type reqSettings struct {
	ProfileKey    string   `json:"profile_key"`
	Description   string   `json:"description"`
	SoulMD        string   `json:"soul_md"`
	EnabledSkills []string `json:"enabled_skills"`
	TenantInfo    string   `json:"tenant_info"`
}

type reqWelcome struct {
	ProfileKey string `json:"profile_key"`
	Pesantren  string `json:"pesantren"`
}

type reqLogHook struct {
	ProfileKey string `json:"profile_key"`
	EventURL   string `json:"event_url"`
	EventKey   string `json:"event_key"`
}

type reqTelegram struct {
	ProfileKey     string  `json:"profile_key"`
	Token          string  `json:"token"`
	Mode           string  `json:"mode"`
	AllowedUserIDs []int64 `json:"allowed_user_ids"`
}

type reqWhatsApp struct {
	ProfileKey     string   `json:"profile_key"`
	Phone          string   `json:"phone"`
	Mode           string   `json:"mode"`
	AllowedNumbers []string `json:"allowed_numbers"`
}

type reqChannelRemove struct {
	ProfileKey string `json:"profile_key"`
	Prefix     string `json:"prefix"`
}

type reqPrune struct {
	ProfileKey string   `json:"profile_key"`
	Names      []string `json:"names"`
}

type reqExtract struct {
	ProfileKey string `json:"profile_key"`
	Name       string `json:"name"`
}

type respExtract struct {
	Made bool `json:"made"`
}

type respReady struct {
	Names []string `json:"names"`
}

type respError struct {
	Error string `json:"error"`
}

// Error codes returned by the bridge; they never carry command output.
const (
	codeInvalidInput = "INVALID_INPUT"
	codeExecFailed   = "EXEC_FAILED"
	codeForbidden    = "FORBIDDEN"
	codeBusy         = "BUSY"
)
