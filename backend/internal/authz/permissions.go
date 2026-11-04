package authz

const (
	Superadmin = "SUPERADMIN"
	Owner      = "OWNER"
	Admin      = "ADMIN"
	Staff      = "STAFF"
	Viewer     = "VIEWER"
)

var rolePermissions = map[string]map[string]bool{
	Superadmin: {"tenant:read": true, "tenant:create": true, "tenant:update": true, "tenant:lifecycle": true, "bot:update": true, "file:read": true, "file:create": true, "file:update": true, "file:delete": true, "bot:read": true, "audit:read": true, "member:manage": true, "platform:manage": true, "assistant:manage": true, "conversation:manage": true},
	Owner:      {"conversation:manage": true, "assistant:manage": true, "tenant:read": true, "tenant:update": true, "file:read": true, "file:create": true, "file:update": true, "file:delete": true, "bot:read": true, "bot:update": true, "audit:read": true, "member:manage": true},
	Admin:      {"conversation:manage": true, "assistant:manage": true, "bot:update": true, "tenant:read": true, "tenant:update": true, "file:read": true, "file:create": true, "file:update": true, "file:delete": true, "bot:read": true, "audit:read": true, "member:manage": true},
	Staff:      {"conversation:manage": true, "tenant:read": true, "file:read": true, "file:create": true, "file:update": true, "bot:read": true},
	Viewer:     {"tenant:read": true, "file:read": true, "bot:read": true},
}

func Permissions(role string) []string {
	out := []string{}
	for p := range rolePermissions[role] {
		out = append(out, p)
	}
	return out
}
func Allows(role, permission string) bool { return rolePermissions[role][permission] }
