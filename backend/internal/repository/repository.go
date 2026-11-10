package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ponpes-saas/internal/contacts"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/questions"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash []byte
	Superadmin   bool
}
type Tenant struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Status       string    `json:"lifecycle_status"`
	Provisioning string    `json:"provisioning_status"`
	CreatedAt    time.Time `json:"created_at"`
}
type File struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	FolderID    *uuid.UUID `json:"folder_id,omitempty"`
	Filename    string     `json:"filename"`
	StorageKey  string     `json:"-"`
	ContentType string     `json:"content_type"`
	Size        int64      `json:"size_bytes"`
	SHA256      string     `json:"sha256"`
	Status      string     `json:"status"`
}
type Folder struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
}

func (r Repository) ListFolders(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID) ([]Folder, error) {
	query := `SELECT id,tenant_id,parent_id,name,created_at FROM folders WHERE tenant_id=$1 AND parent_id IS NOT DISTINCT FROM $2 ORDER BY name`
	rows, err := r.DB.Query(ctx, query, tenantID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Folder{}
	for rows.Next() {
		var item Folder
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ParentID, &item.Name, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (r Repository) CreateFolder(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, name string) (Folder, error) {
	var item Folder
	err := r.DB.QueryRow(ctx, `INSERT INTO folders(tenant_id,parent_id,name) VALUES($1,$2,$3) RETURNING id,tenant_id,parent_id,name,created_at`, tenantID, parentID, name).Scan(&item.ID, &item.TenantID, &item.ParentID, &item.Name, &item.CreatedAt)
	return item, err
}

type Profile struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Website     string    `json:"website"`
	// GoogleMapsURL is the pesantren's official location link. It is stored and shared, never fetched by the server.
	GoogleMapsURL string `json:"google_maps_url"`
	// Social pages shown on the public link page. Empty means none.
	InstagramURL string `json:"instagram_url"`
	FacebookURL  string `json:"facebook_url"`
	TiktokURL    string `json:"tiktok_url"`
	YoutubeURL   string `json:"youtube_url"`
	// HasLogo is read-only: the logo is managed through the logo endpoints, never through the profile form.
	HasLogo bool `json:"has_logo"`
}

type TelegramBot struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	BotID         int64     `json:"telegram_bot_id"`
	Username      string    `json:"username"`
	CredentialRef string    `json:"credential_ref"`
	Status        string    `json:"status"`
	// AccessMode is TESTING (only AllowedUserIDs may chat) or PUBLIC (anyone may chat).
	AccessMode     string  `json:"access_mode"`
	AllowedUserIDs []int64 `json:"allowed_user_ids"`
}

// WhatsAppChannel is a tenant's linked WhatsApp number. It mirrors TelegramBot: same access modes and statuses.
type WhatsAppChannel struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	PhoneNumber string    `json:"phone_number"`
	Status      string    `json:"status"`
	// AccessMode is TESTING (only AllowedNumbers may chat) or PUBLIC (anyone may chat).
	AccessMode     string   `json:"access_mode"`
	AllowedNumbers []string `json:"allowed_numbers"`
	// ServiceMode is PSB_ONLY (a dedicated admissions line) or GENERAL (a shared line that asks what the visitor needs).
	ServiceMode string `json:"service_mode"`
}

type HermesSettings struct {
	Description   string    `json:"description"`
	SoulMD        string    `json:"soul_md"`
	EnabledSkills []string  `json:"enabled_skills"`
	Revision      int       `json:"config_revision"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Repository struct{ DB *pgxpool.Pool }

type Membership struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       string    `json:"role"`
}
type Member struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       string    `json:"role"`
}

type AuditLog struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   *uuid.UUID `json:"tenant_id,omitempty"`
	ActorType  string     `json:"actor_type"`
	ActorName  string     `json:"actor_name"`
	Action     string     `json:"action"`
	TargetType string     `json:"target_type,omitempty"`
	TargetID   string     `json:"target_id,omitempty"`
	TargetName string     `json:"target_name,omitempty"`
	OccurredAt time.Time  `json:"occurred_at"`
}

type Conversation struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	TelegramChatID int64     `json:"telegram_chat_id"`
	TelegramUserID *int64    `json:"telegram_user_id,omitempty"`
	Username       string    `json:"username,omitempty"`
	Status         string    `json:"status"`
	StartedAt      time.Time `json:"started_at"`
	LastMessageAt  time.Time `json:"last_message_at"`
	// Snippet is the newest message matching a search query; empty when no query was given.
	Snippet string `json:"snippet,omitempty"`
}

type TelegramMessage struct {
	ID                uuid.UUID `json:"id"`
	ConversationID    uuid.UUID `json:"conversation_id"`
	SenderType        string    `json:"sender_type"`
	MessageText       string    `json:"message_text"`
	TelegramMessageID *int64    `json:"telegram_message_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type TelegramExportRow struct {
	ConversationID uuid.UUID
	Username       string
	ChatID         int64
	SenderType     string
	MessageText    string
	CreatedAt      time.Time
}

// TelegramExportFilter limits an Excel export to chosen conversations. Dates are applied after question and answer are paired.
type TelegramExportFilter struct {
	ConversationIDs []uuid.UUID
}

func (r Repository) TelegramExportRows(ctx context.Context, tenantID uuid.UUID, filter TelegramExportFilter) ([]TelegramExportRow, error) {
	args := []any{tenantID}
	conditions := []string{`c.tenant_id=$1`}
	if len(filter.ConversationIDs) > 0 {
		args = append(args, filter.ConversationIDs)
		conditions = append(conditions, `c.id = ANY($`+strconv.Itoa(len(args))+`)`)
	}
	rows, err := r.DB.Query(ctx, `SELECT m.conversation_id,coalesce(c.username,''),c.telegram_chat_id,m.sender_type,m.message_text,m.created_at FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id WHERE `+strings.Join(conditions, ` AND `)+` ORDER BY m.created_at ASC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TelegramExportRow{}
	for rows.Next() {
		var item TelegramExportRow
		if err := rows.Scan(&item.ConversationID, &item.Username, &item.ChatID, &item.SenderType, &item.MessageText, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// likePattern escapes LIKE wildcards so a search term is matched literally.
func likePattern(term string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return "%" + replacer.Replace(term) + "%"
}

func (r Repository) ConversationMessages(ctx context.Context, conversationID uuid.UUID) ([]TelegramMessage, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,conversation_id,sender_type,message_text,telegram_message_id,created_at FROM telegram_messages WHERE conversation_id=$1 ORDER BY created_at ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TelegramMessage{}
	for rows.Next() {
		var item TelegramMessage
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.SenderType, &item.MessageText, &item.TelegramMessageID, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// Conversations lists a tenant's conversations. A non-empty search matches the username or any message text.
func (r Repository) Conversations(ctx context.Context, tenantID *uuid.UUID, search string) ([]Conversation, error) {
	snippet := `''`
	conditions := []string{}
	args := []any{}
	if tenantID != nil {
		args = append(args, *tenantID)
		conditions = append(conditions, `c.tenant_id=$`+strconv.Itoa(len(args)))
	}
	if search = strings.TrimSpace(search); search != "" {
		args = append(args, likePattern(search))
		n := strconv.Itoa(len(args))
		match := `SELECT 1 FROM telegram_messages m WHERE m.conversation_id=c.id AND m.message_text ILIKE $` + n
		snippet = `coalesce((SELECT left(m.message_text,160) FROM telegram_messages m WHERE m.conversation_id=c.id AND m.message_text ILIKE $` + n + ` ORDER BY m.created_at DESC LIMIT 1),'')`
		conditions = append(conditions, `(coalesce(c.username,'') ILIKE $`+n+` OR EXISTS (`+match+`))`)
	}
	query := `SELECT c.id,c.tenant_id,c.telegram_chat_id,c.telegram_user_id,coalesce(c.username,''),c.status,c.started_at,c.last_message_at,` + snippet + ` FROM telegram_conversations c`
	if len(conditions) > 0 {
		query += ` WHERE ` + strings.Join(conditions, ` AND `)
	}
	query += ` ORDER BY c.last_message_at DESC LIMIT 200`
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Conversation{}
	for rows.Next() {
		var item Conversation
		if err := rows.Scan(&item.ID, &item.TenantID, &item.TelegramChatID, &item.TelegramUserID, &item.Username, &item.Status, &item.StartedAt, &item.LastMessageAt, &item.Snippet); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ConversationGroup is a conversation with the question counters shown in the grouped Telegram log.
type ConversationGroup struct {
	Conversation
	Turns        int    `json:"turns"`
	Unanswered   int    `json:"unanswered"`
	LastQuestion string `json:"last_question"`
}

// ConversationGroupQuery filters, sorts and pages a tenant's grouped Telegram log.
type ConversationGroupQuery struct {
	Search string
	Status string
	Sort   string // recent (default), most, unanswered
	Limit  int
	Offset int
}

// ConversationGroups returns one page of a tenant's conversations with question counts, the total match count and every status in use.
// A question is unanswered when the next message in its conversation is not from the bot.
func (r Repository) ConversationGroups(ctx context.Context, tenantID uuid.UUID, q ConversationGroupQuery) ([]ConversationGroup, int, []string, error) {
	args := []any{tenantID}
	conditions := []string{`c.tenant_id=$1`}
	if q.Status != "" {
		args = append(args, q.Status)
		conditions = append(conditions, `c.status=$`+strconv.Itoa(len(args)))
	}
	if search := strings.TrimSpace(q.Search); search != "" {
		args = append(args, likePattern(search))
		n := strconv.Itoa(len(args))
		conditions = append(conditions, `(coalesce(c.username,'') ILIKE $`+n+` OR c.telegram_chat_id::text ILIKE $`+n+` OR EXISTS (SELECT 1 FROM telegram_messages m WHERE m.conversation_id=c.id AND m.sender_type='USER' AND m.message_text ILIKE $`+n+`))`)
	}
	order := `c.last_message_at DESC`
	switch q.Sort {
	case "most":
		order = `coalesce(s.turns,0) DESC, c.last_message_at DESC`
	case "unanswered":
		order = `coalesce(s.unanswered,0) DESC, c.last_message_at DESC`
	}
	args = append(args, q.Limit, q.Offset)
	limit, offset := strconv.Itoa(len(args)-1), strconv.Itoa(len(args))
	query := `WITH msgs AS (
SELECT m.conversation_id, m.sender_type, m.message_text, m.created_at,
LEAD(m.sender_type) OVER (PARTITION BY m.conversation_id ORDER BY m.created_at, m.id) AS next_sender
FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id WHERE c.tenant_id=$1),
s AS (SELECT conversation_id,
count(*) FILTER (WHERE sender_type='USER') AS turns,
count(*) FILTER (WHERE sender_type='USER' AND coalesce(next_sender,'')<>'BOT') AS unanswered,
(array_agg(message_text ORDER BY created_at DESC) FILTER (WHERE sender_type='USER'))[1] AS last_question
FROM msgs GROUP BY conversation_id)
SELECT c.id,c.tenant_id,c.telegram_chat_id,c.telegram_user_id,coalesce(c.username,''),c.status,c.started_at,c.last_message_at,
coalesce(s.turns,0),coalesce(s.unanswered,0),coalesce(left(s.last_question,200),''),count(*) OVER ()
FROM telegram_conversations c LEFT JOIN s ON s.conversation_id=c.id
WHERE ` + strings.Join(conditions, ` AND `) + ` ORDER BY ` + order + ` LIMIT $` + limit + ` OFFSET $` + offset
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, nil, err
	}
	defer rows.Close()
	out := []ConversationGroup{}
	total := 0
	for rows.Next() {
		var item ConversationGroup
		if err := rows.Scan(&item.ID, &item.TenantID, &item.TelegramChatID, &item.TelegramUserID, &item.Username, &item.Status, &item.StartedAt, &item.LastMessageAt, &item.Turns, &item.Unanswered, &item.LastQuestion, &total); err != nil {
			return nil, 0, nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, nil, err
	}
	statusRows, err := r.DB.Query(ctx, `SELECT DISTINCT status FROM telegram_conversations WHERE tenant_id=$1 ORDER BY status`, tenantID)
	if err != nil {
		return nil, 0, nil, err
	}
	defer statusRows.Close()
	statuses := []string{}
	for statusRows.Next() {
		var status string
		if err := statusRows.Scan(&status); err != nil {
			return nil, 0, nil, err
		}
		statuses = append(statuses, status)
	}
	return out, total, statuses, statusRows.Err()
}

// SetConversationStatus moves a conversation between ACTIVE, HANDOFF and RESOLVED.
func (r Repository) SetConversationStatus(ctx context.Context, conversationID uuid.UUID, status string) (Conversation, error) {
	var item Conversation
	err := r.DB.QueryRow(ctx, `UPDATE telegram_conversations SET status=$2, closed_at=CASE WHEN $2='RESOLVED' THEN now() ELSE NULL END WHERE id=$1 RETURNING id,tenant_id,telegram_chat_id,telegram_user_id,coalesce(username,''),status,started_at,last_message_at`, conversationID, status).Scan(&item.ID, &item.TenantID, &item.TelegramChatID, &item.TelegramUserID, &item.Username, &item.Status, &item.StartedAt, &item.LastMessageAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}

// TouchConversation marks a conversation as active now without changing its status.
func (r Repository) TouchConversation(ctx context.Context, conversationID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `UPDATE telegram_conversations SET last_message_at=now() WHERE id=$1`, conversationID)
	return err
}

func (r Repository) ConversationTenant(ctx context.Context, conversationID uuid.UUID) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := r.DB.QueryRow(ctx, `SELECT tenant_id FROM telegram_conversations WHERE id=$1`, conversationID).Scan(&tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return tenantID, ErrNotFound
	}
	return tenantID, err
}

func (r Repository) UpsertConversation(ctx context.Context, tenantID uuid.UUID, channel string, chatID, userID int64, username string) (Conversation, error) {
	var item Conversation
	err := r.DB.QueryRow(ctx, `INSERT INTO telegram_conversations(tenant_id,channel,telegram_chat_id,telegram_user_id,username,last_message_at) VALUES($1,$5,$2,$3,$4,now()) ON CONFLICT (tenant_id,channel,telegram_chat_id) DO UPDATE SET telegram_user_id=EXCLUDED.telegram_user_id,username=EXCLUDED.username,last_message_at=now(),status=CASE WHEN telegram_conversations.status='HANDOFF' THEN 'HANDOFF' ELSE 'ACTIVE' END RETURNING id,tenant_id,telegram_chat_id,telegram_user_id,coalesce(username,''),status,started_at,last_message_at`, tenantID, chatID, userID, username, channel).Scan(&item.ID, &item.TenantID, &item.TelegramChatID, &item.TelegramUserID, &item.Username, &item.Status, &item.StartedAt, &item.LastMessageAt)
	return item, err
}

// FindConversation returns an existing conversation without creating one (ErrNotFound when there is none).
func (r Repository) FindConversation(ctx context.Context, tenantID uuid.UUID, channel string, chatID int64) (Conversation, error) {
	var item Conversation
	err := r.DB.QueryRow(ctx, `SELECT id,tenant_id,telegram_chat_id,telegram_user_id,coalesce(username,''),status,started_at,last_message_at FROM telegram_conversations WHERE tenant_id=$1 AND channel=$2 AND telegram_chat_id=$3`, tenantID, channel, chatID).Scan(&item.ID, &item.TenantID, &item.TelegramChatID, &item.TelegramUserID, &item.Username, &item.Status, &item.StartedAt, &item.LastMessageAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}

func (r Repository) AddTelegramMessage(ctx context.Context, conversationID uuid.UUID, senderType, text string, telegramMessageID *int64) (TelegramMessage, error) {
	var item TelegramMessage
	err := r.DB.QueryRow(ctx, `INSERT INTO telegram_messages(conversation_id,sender_type,message_text,telegram_message_id) VALUES($1,$2,$3,$4) RETURNING id,conversation_id,sender_type,message_text,telegram_message_id,created_at`, conversationID, senderType, text, telegramMessageID).Scan(&item.ID, &item.ConversationID, &item.SenderType, &item.MessageText, &item.TelegramMessageID, &item.CreatedAt)
	return item, err
}

// RecordAudit writes a best-effort audit entry; a failure must never break the action being audited.
// name is a human-readable target label (e.g. a filename) shown in the audit log.
func (r Repository) RecordAudit(ctx context.Context, tenantID uuid.UUID, actor *uuid.UUID, action, targetType, targetID, name string) {
	actorType := "USER"
	if actor == nil {
		actorType = "SYSTEM"
	}
	details, _ := json.Marshal(map[string]string{"name": name})
	_, _ = r.DB.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id,details_redacted) VALUES($1,$2,$3,$4,$5,$6,$7)`, tenantID, actor, actorType, action, targetType, targetID, details)
}

// RecordGlobalAudit records a platform-wide action that belongs to no tenant (tenant_id stays NULL).
func (r Repository) RecordGlobalAudit(ctx context.Context, actor *uuid.UUID, action, targetType, targetID string) {
	actorType := "USER"
	if actor == nil {
		actorType = "SYSTEM"
	}
	_, _ = r.DB.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES(NULL,$1,$2,$3,$4,$5)`, actor, actorType, action, targetType, targetID)
}

func (r Repository) AuditLogs(ctx context.Context, tenantID *uuid.UUID) ([]AuditLog, error) {
	query := `SELECT a.id,a.tenant_id,a.actor_type,coalesce(actor.email,''),a.action,coalesce(a.target_type,''),coalesce(a.target_id,''),coalesce(nullif(a.details_redacted->>'name',''),target_user.email,target_tenant.name,''),a.occurred_at FROM audit_logs a LEFT JOIN users actor ON actor.id=a.actor_user_id LEFT JOIN users target_user ON a.target_type='USER' AND target_user.id=CASE WHEN a.target_id ~ '^[0-9a-f-]{36}$' THEN a.target_id::uuid ELSE NULL END LEFT JOIN tenants target_tenant ON a.target_type='TENANT' AND target_tenant.id=CASE WHEN a.target_id ~ '^[0-9a-f-]{36}$' THEN a.target_id::uuid ELSE NULL END`
	args := []any{}
	if tenantID != nil {
		query += ` WHERE tenant_id=$1`
		args = append(args, *tenantID)
	}
	query += ` ORDER BY occurred_at DESC LIMIT 200`
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditLog{}
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ActorType, &item.ActorName, &item.Action, &item.TargetType, &item.TargetID, &item.TargetName, &item.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r Repository) TenantMembers(ctx context.Context, tenantID uuid.UUID) ([]Member, error) {
	rows, err := r.DB.Query(ctx, `SELECT u.id,u.email,tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN users u ON u.id=tu.user_id JOIN tenants t ON t.id=tu.tenant_id WHERE tu.tenant_id=$1 ORDER BY tu.role,u.email`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var item Member
		if err := rows.Scan(&item.UserID, &item.Email, &item.TenantID, &item.TenantName, &item.Role); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r Repository) ListMemberships(ctx context.Context, tenantID uuid.UUID) ([]Membership, error) {
	rows, err := r.DB.Query(ctx, `SELECT tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN tenants t ON t.id=tu.tenant_id WHERE tu.tenant_id=$1 ORDER BY t.name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Membership{}
	for rows.Next() {
		var m Membership
		if err := rows.Scan(&m.TenantID, &m.TenantName, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r Repository) ListAuditLogs(ctx context.Context, tenantID *uuid.UUID) ([]AuditLog, error) {
	query := `SELECT id,tenant_id,actor_type,action,coalesce(target_type,''),occurred_at FROM audit_logs`
	args := []any{}
	if tenantID != nil {
		query += ` WHERE tenant_id=$1`
		args = append(args, *tenantID)
	}
	query += ` ORDER BY occurred_at DESC LIMIT 200`
	rows, err := r.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditLog{}
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(&item.ID, &item.TenantID, &item.ActorType, &item.Action, &item.TargetType, &item.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r Repository) HermesOverview(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.DB.Query(ctx, `SELECT t.id,t.name,coalesce(h.runtime_status,'UNKNOWN'),(SELECT count(*) FROM files f WHERE f.tenant_id=t.id AND f.deleted_at IS NULL) FROM tenants t LEFT JOIN hermes_profiles h ON h.tenant_id=t.id ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id uuid.UUID
		var name, status string
		var files int64
		if err := rows.Scan(&id, &name, &status, &files); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"tenant_id": id, "tenant_name": name, "runtime_status": status, "file_count": files})
	}
	return out, rows.Err()
}

type HermesOverview struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	TenantName    string    `json:"tenant_name"`
	RuntimeStatus string    `json:"runtime_status"`
	FileCount     int       `json:"file_count"`
}

func (r Repository) Membership(ctx context.Context, userID, tenantID uuid.UUID) (Membership, error) {
	var m Membership
	err := r.DB.QueryRow(ctx, `SELECT tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN tenants t ON t.id=tu.tenant_id WHERE tu.user_id=$1 AND tu.tenant_id=$2`, userID, tenantID).Scan(&m.TenantID, &m.TenantName, &m.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNotFound
	}
	return m, err
}

func (r Repository) TenantRole(ctx context.Context, userID, tenantID uuid.UUID) (string, error) {
	var role string
	err := r.DB.QueryRow(ctx, `SELECT role FROM tenant_users WHERE user_id=$1 AND tenant_id=$2`, userID, tenantID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (r Repository) Memberships(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
	rows, err := r.DB.Query(ctx, `SELECT tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN tenants t ON t.id=tu.tenant_id WHERE tu.user_id=$1 ORDER BY t.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Membership{}
	for rows.Next() {
		var m Membership
		if err := rows.Scan(&m.TenantID, &m.TenantName, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r Repository) ListMemberTenants(ctx context.Context, userID uuid.UUID) ([]Tenant, error) {
	rows, err := r.DB.Query(ctx, `SELECT t.id,t.name,t.slug,t.lifecycle_status,coalesce(pr.status,'NOT_STARTED'),t.created_at FROM tenants t JOIN tenant_users tu ON tu.tenant_id=t.id LEFT JOIN LATERAL (SELECT status FROM provisioning_runs WHERE tenant_id=t.id ORDER BY created_at DESC LIMIT 1) pr ON true WHERE tu.user_id=$1 ORDER BY t.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tenant{}
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Provisioning, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r Repository) FindUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	var u User
	err := r.DB.QueryRow(ctx, `SELECT id,email,password_hash,is_platform_superadmin FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Superadmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}
func (r Repository) FindUser(ctx context.Context, email string) (User, error) {
	var u User
	err := r.DB.QueryRow(ctx, `SELECT id,email,password_hash,is_platform_superadmin FROM users WHERE lower(email)=lower($1)`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Superadmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}
func (r Repository) CreateSession(ctx context.Context, userID uuid.UUID, raw string, expires time.Time) error {
	sum := sha256.Sum256([]byte(raw))
	_, err := r.DB.Exec(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)`, hex.EncodeToString(sum[:]), userID, expires)
	return err
}
func (r Repository) SessionUser(ctx context.Context, raw string) (uuid.UUID, error) {
	sum := sha256.Sum256([]byte(raw))
	var id uuid.UUID
	err := r.DB.QueryRow(ctx, `SELECT user_id FROM sessions WHERE token_hash=$1 AND expires_at>now()`, hex.EncodeToString(sum[:])).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return id, ErrNotFound
	}
	return id, err
}
func (r Repository) DeleteSession(ctx context.Context, raw string) error {
	sum := sha256.Sum256([]byte(raw))
	_, err := r.DB.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, hex.EncodeToString(sum[:]))
	return err
}
func (r Repository) UpdateTenant(ctx context.Context, id uuid.UUID, name, slug string, actor uuid.UUID) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE tenants SET name=$2,slug=$3,updated_at=now() WHERE id=$1`, id, name, slug)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES($1,$2,'USER','TENANT_UPDATED','TENANT',$3)`, id, actor, id.String()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r Repository) CreateMember(ctx context.Context, tenantID uuid.UUID, email string, password []byte, role string) (uuid.UUID, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var id uuid.UUID
	if err = tx.QueryRow(ctx, `INSERT INTO users(email,password_hash) VALUES(lower($1),$2) RETURNING id`, email, password).Scan(&id); err != nil {
		return uuid.Nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO tenant_users(tenant_id,user_id,role) VALUES($1,$2,$3)`, tenantID, id, role); err != nil {
		return uuid.Nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES($1,NULL,'SYSTEM','MEMBER_CREATED','USER',$2)`, tenantID, id.String()); err != nil {
		return uuid.Nil, err
	}
	return id, tx.Commit(ctx)
}
func (r Repository) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, role string, actor uuid.UUID) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE tenant_users SET role=$3 WHERE tenant_id=$1 AND user_id=$2`, tenantID, userID, role)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES($1,$2,'USER','MEMBER_ROLE_UPDATED','USER',$3)`, tenantID, actor, userID.String()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r Repository) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]Member, error) {
	rows, err := r.DB.Query(ctx, `SELECT u.id,u.email,tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN users u ON u.id=tu.user_id JOIN tenants t ON t.id=tu.tenant_id WHERE tu.tenant_id=$1 ORDER BY tu.role,u.email`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Email, &m.TenantID, &m.TenantName, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r Repository) CreateTenant(ctx context.Context, name, slug string, actor uuid.UUID) (Tenant, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return Tenant{}, err
	}
	defer tx.Rollback(ctx)
	var t Tenant
	err = tx.QueryRow(ctx, `INSERT INTO tenants(name,slug,lifecycle_status) VALUES($1,$2,'DRAFT') RETURNING id,name,slug,lifecycle_status,created_at`, name, slug).Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.CreatedAt)
	if err != nil {
		return t, err
	}
	t.Provisioning = "NOT_STARTED"
	var actorValue any = actor
	if actor == uuid.Nil {
		actorValue = nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES($1,$2,'USER','TENANT_CREATED','TENANT',$3)`, t.ID, actorValue, t.ID.String())
	if err != nil {
		return t, err
	}
	return t, tx.Commit(ctx)
}
func (r Repository) QueueHermesProvisioning(ctx context.Context, tenantID uuid.UUID) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE tenants SET lifecycle_status='ONBOARDING',updated_at=now() WHERE id=$1 AND NOT EXISTS (SELECT 1 FROM hermes_profiles WHERE tenant_id=$1)`, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO provisioning_runs(tenant_id,kind,status) VALUES($1,'ONBOARDING','QUEUED')`, tenantID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r Repository) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := r.DB.Query(ctx, `SELECT t.id,t.name,t.slug,t.lifecycle_status,coalesce(pr.status,'NOT_STARTED'),t.created_at FROM tenants t LEFT JOIN LATERAL (SELECT status FROM provisioning_runs WHERE tenant_id=t.id ORDER BY created_at DESC LIMIT 1) pr ON true ORDER BY t.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tenant{}
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Provisioning, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r Repository) ListTenantsForUser(ctx context.Context, userID uuid.UUID) ([]Tenant, error) {
	rows, err := r.DB.Query(ctx, `SELECT t.id,t.name,t.slug,t.lifecycle_status,coalesce(pr.status,'NOT_STARTED'),t.created_at FROM tenants t JOIN tenant_users tu ON tu.tenant_id=t.id LEFT JOIN LATERAL (SELECT status FROM provisioning_runs WHERE tenant_id=t.id ORDER BY created_at DESC LIMIT 1) pr ON true WHERE tu.user_id=$1 ORDER BY t.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tenant{}
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Provisioning, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (r Repository) GetTenant(ctx context.Context, id uuid.UUID) (Tenant, error) {
	var t Tenant
	err := r.DB.QueryRow(ctx, `SELECT t.id,t.name,t.slug,t.lifecycle_status,coalesce(pr.status,'NOT_STARTED'),t.created_at FROM tenants t LEFT JOIN LATERAL (SELECT status FROM provisioning_runs WHERE tenant_id=t.id ORDER BY created_at DESC LIMIT 1) pr ON true WHERE t.id=$1`, id).Scan(&t.ID, &t.Name, &t.Slug, &t.Status, &t.Provisioning, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (r Repository) Members(ctx context.Context, tenantID uuid.UUID) ([]Membership, error) {
	rows, err := r.DB.Query(ctx, `SELECT tu.tenant_id,t.name,tu.role FROM tenant_users tu JOIN tenants t ON t.id=tu.tenant_id WHERE tu.tenant_id=$1 ORDER BY tu.role`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Membership{}
	for rows.Next() {
		var member Membership
		if err := rows.Scan(&member.TenantID, &member.TenantName, &member.Role); err != nil {
			return nil, err
		}
		out = append(out, member)
	}
	return out, rows.Err()
}

func (r Repository) GetProfile(ctx context.Context, tenantID uuid.UUID) (Profile, error) {
	var p Profile
	err := r.DB.QueryRow(ctx, `SELECT tenant_id,display_name,coalesce(description,''),coalesce(address,''),coalesce(phone,''),coalesce(email,''),coalesce(website,''),coalesce(google_maps_url,''),coalesce(instagram_url,''),coalesce(facebook_url,''),coalesce(tiktok_url,''),coalesce(youtube_url,''),logo_key IS NOT NULL FROM tenant_profiles WHERE tenant_id=$1`, tenantID).Scan(&p.TenantID, &p.DisplayName, &p.Description, &p.Address, &p.Phone, &p.Email, &p.Website, &p.GoogleMapsURL, &p.InstagramURL, &p.FacebookURL, &p.TiktokURL, &p.YoutubeURL, &p.HasLogo)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (r Repository) UpsertProfile(ctx context.Context, p Profile, actor uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO tenant_profiles(tenant_id,display_name,description,address,phone,email,website,google_maps_url,instagram_url,facebook_url,tiktok_url,youtube_url,updated_by,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,NULLIF($9,''),NULLIF($10,''),NULLIF($11,''),NULLIF($12,''),NULLIF($13,''),$8,now()) ON CONFLICT (tenant_id) DO UPDATE SET display_name=EXCLUDED.display_name,description=EXCLUDED.description,address=EXCLUDED.address,phone=EXCLUDED.phone,email=EXCLUDED.email,website=EXCLUDED.website,google_maps_url=EXCLUDED.google_maps_url,instagram_url=EXCLUDED.instagram_url,facebook_url=EXCLUDED.facebook_url,tiktok_url=EXCLUDED.tiktok_url,youtube_url=EXCLUDED.youtube_url,updated_by=EXCLUDED.updated_by,updated_at=now()`, p.TenantID, p.DisplayName, p.Description, p.Address, p.Phone, p.Email, p.Website, actor, strings.TrimSpace(p.GoogleMapsURL), p.InstagramURL, p.FacebookURL, p.TiktokURL, p.YoutubeURL)
	return err
}

func (r Repository) GetTelegramBot(ctx context.Context, tenantID uuid.UUID) (TelegramBot, error) {
	var b TelegramBot
	err := r.DB.QueryRow(ctx, `SELECT id,tenant_id,telegram_bot_id,username,credential_ref,status,access_mode,allowed_user_ids FROM telegram_bots WHERE tenant_id=$1`, tenantID).Scan(&b.ID, &b.TenantID, &b.BotID, &b.Username, &b.CredentialRef, &b.Status, &b.AccessMode, &b.AllowedUserIDs)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (r Repository) UpsertTelegramBot(ctx context.Context, b TelegramBot, actor uuid.UUID) error {
	if b.AllowedUserIDs == nil {
		b.AllowedUserIDs = []int64{}
	}
	_, err := r.DB.Exec(ctx, `INSERT INTO telegram_bots(tenant_id,telegram_bot_id,username,credential_ref,status,access_mode,allowed_user_ids) VALUES($1,$2,$3,$4,'PENDING',$5,$6) ON CONFLICT (tenant_id) DO UPDATE SET telegram_bot_id=EXCLUDED.telegram_bot_id,username=EXCLUDED.username,credential_ref=EXCLUDED.credential_ref,access_mode=EXCLUDED.access_mode,allowed_user_ids=EXCLUDED.allowed_user_ids,status='PENDING',updated_at=now()`, b.TenantID, b.BotID, b.Username, b.CredentialRef, b.AccessMode, b.AllowedUserIDs)
	return err
}

func (r Repository) SetTelegramBotStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	_, err := r.DB.Exec(ctx, `UPDATE telegram_bots SET status=$2, verified_at=CASE WHEN $2='ACTIVE' THEN now() ELSE verified_at END, updated_at=now() WHERE tenant_id=$1`, tenantID, status)
	return err
}

func (r Repository) DeleteTelegramBot(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM telegram_bots WHERE tenant_id=$1`, tenantID)
	return err
}

func (r Repository) InsertFile(ctx context.Context, f File) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO files(id,tenant_id,folder_id,filename,storage_key,content_type,size_bytes,sha256,processing_status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, f.ID, f.TenantID, f.FolderID, f.Filename, f.StorageKey, f.ContentType, f.Size, f.SHA256, f.Status)
	return err
}
func (r Repository) ListFiles(ctx context.Context, tenantID uuid.UUID) ([]File, error) {
	rows, err := r.DB.Query(ctx, `SELECT id,tenant_id,folder_id,filename,storage_key,content_type,size_bytes,sha256,processing_status FROM files WHERE tenant_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []File{}
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.TenantID, &f.FolderID, &f.Filename, &f.StorageKey, &f.ContentType, &f.Size, &f.SHA256, &f.Status); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r Repository) MarkFilesReady(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `UPDATE files SET processing_status='READY', updated_at=now() WHERE tenant_id=$1 AND processing_status='PENDING_SYNC' AND deleted_at IS NULL`, tenantID)
	return err
}

func (r Repository) HermesProfileKey(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var key string
	if err := r.DB.QueryRow(ctx, `SELECT profile_key FROM hermes_profiles WHERE tenant_id=$1`, tenantID).Scan(&key); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return key, nil
}

func (r Repository) TenantByHermesProfileKey(ctx context.Context, profileKey string) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := r.DB.QueryRow(ctx, `SELECT tenant_id FROM hermes_profiles WHERE profile_key=$1`, profileKey).Scan(&tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return tenantID, ErrNotFound
	}
	return tenantID, err
}

// ProfileInfoByKey renders the tenant profile as short markdown for the assistant's persona.
// It returns "" when the tenant has no profile yet.
// PesantrenNameByKey is the public name of the pesantren behind a Hermes profile (profile display name, else tenant name).
func (r Repository) PesantrenNameByKey(ctx context.Context, profileKey string) string {
	var name string
	err := r.DB.QueryRow(ctx, `SELECT coalesce(nullif(p.display_name,''),t.name) FROM hermes_profiles h JOIN tenants t ON t.id=h.tenant_id LEFT JOIN tenant_profiles p ON p.tenant_id=h.tenant_id WHERE h.profile_key=$1`, profileKey).Scan(&name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

func (r Repository) ProfileInfoByKey(ctx context.Context, profileKey string) string {
	var name, description, address, phone, email, website, maps string
	err := r.DB.QueryRow(ctx, `SELECT coalesce(p.display_name,''),coalesce(p.description,''),coalesce(p.address,''),coalesce(p.phone,''),coalesce(p.email,''),coalesce(p.website,''),coalesce(p.google_maps_url,'') FROM hermes_profiles h JOIN tenant_profiles p ON p.tenant_id=h.tenant_id WHERE h.profile_key=$1`, profileKey).Scan(&name, &description, &address, &phone, &email, &website, &maps)
	if err != nil {
		return ""
	}
	out := ""
	for _, item := range [][2]string{{"Nama pesantren", name}, {"Tentang", description}, {"Alamat dan informasi layanan", address}, {"Telepon/WhatsApp", phone}, {"Email", email}, {"Website", website}, {"Link Google Maps (lokasi resmi)", maps}} {
		if strings.TrimSpace(item[1]) != "" {
			out += "- " + item[0] + ": " + strings.TrimSpace(item[1]) + "\n"
		}
	}
	return out + links.PersonaSection(r.SharedLinksByProfileKey(ctx, profileKey)) + contacts.PersonaSection(r.SharedContactsByProfileKey(ctx, profileKey)) + questions.PersonaSection(r.AnsweredQuestionsByProfileKey(ctx, profileKey))
}

// HermesProfileRef identifies one prepared assistant together with its pesantren.
type HermesProfileRef struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	ProfileKey string    `json:"profile_key"`
}

func (r Repository) HermesProfiles(ctx context.Context) ([]HermesProfileRef, error) {
	rows, err := r.DB.Query(ctx, `SELECT h.tenant_id,t.name,h.profile_key FROM hermes_profiles h JOIN tenants t ON t.id=h.tenant_id ORDER BY h.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []HermesProfileRef{}
	for rows.Next() {
		var item HermesProfileRef
		if err := rows.Scan(&item.TenantID, &item.TenantName, &item.ProfileKey); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r Repository) HermesProfileKeys(ctx context.Context) ([]string, error) {
	rows, err := r.DB.Query(ctx, `SELECT profile_key FROM hermes_profiles ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (r Repository) DeleteHermesProfile(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM hermes_profiles WHERE tenant_id=$1`, tenantID)
	return err
}

func (r Repository) GetGlobalHermesSettings(ctx context.Context) (HermesSettings, error) {
	var settings HermesSettings
	err := r.DB.QueryRow(ctx, `SELECT description,soul_md,enabled_skills,config_revision,updated_at FROM hermes_global_settings WHERE id=1`).Scan(&settings.Description, &settings.SoulMD, &settings.EnabledSkills, &settings.Revision, &settings.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return HermesSettings{EnabledSkills: []string{}, Revision: 1}, nil
	}
	return settings, err
}

func (r Repository) SaveGlobalHermesSettings(ctx context.Context, settings HermesSettings, actor uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO hermes_global_settings(id,description,soul_md,enabled_skills,config_revision,updated_by,updated_at) VALUES(1,$1,$2,$3,1,$4,now()) ON CONFLICT (id) DO UPDATE SET description=EXCLUDED.description,soul_md=EXCLUDED.soul_md,enabled_skills=EXCLUDED.enabled_skills,config_revision=hermes_global_settings.config_revision+1,updated_by=EXCLUDED.updated_by,updated_at=now()`, settings.Description, settings.SoulMD, settings.EnabledSkills, actor)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES(NULL,$1,'USER','HERMES_GLOBAL_SETTINGS_UPDATED','HERMES_GLOBAL_SETTINGS','global')`, actor)
	return err
}
func (r Repository) FindFile(ctx context.Context, tenantID, fileID uuid.UUID) (File, error) {
	var f File
	err := r.DB.QueryRow(ctx, `SELECT id,tenant_id,folder_id,filename,storage_key,content_type,size_bytes,sha256,processing_status FROM files WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, fileID, tenantID).Scan(&f.ID, &f.TenantID, &f.FolderID, &f.Filename, &f.StorageKey, &f.ContentType, &f.Size, &f.SHA256, &f.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return f, ErrNotFound
	}
	return f, err
}
func (r Repository) DeleteFile(ctx context.Context, tenantID, fileID uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `UPDATE files SET deleted_at=now(), updated_at=now() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, fileID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (r Repository) RenameFile(ctx context.Context, tenantID, fileID uuid.UUID, filename string) error {
	tag, err := r.DB.Exec(ctx, `UPDATE files SET filename=$3, updated_at=now() WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`, fileID, tenantID, filename)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
func (r Repository) SetTenantStatus(ctx context.Context, id uuid.UUID, status string, actor uuid.UUID) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE tenants SET lifecycle_status=$2,updated_at=now() WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_logs(tenant_id,actor_user_id,actor_type,action,target_type,target_id) VALUES($1,$2,'USER',$3,'TENANT',$4)`, id, actor, "TENANT_"+status, id.String()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ConversationStats summarises Telegram activity for the dashboard. Days and hours use the Asia/Jakarta calendar.
type ConversationStats struct {
	Days            int             `json:"days"`
	UserMessages    int             `json:"user_messages"`
	BotReplies      int             `json:"bot_replies"`
	PrevUser        int             `json:"prev_user_messages"`
	PrevBot         int             `json:"prev_bot_replies"`
	Visitors        int             `json:"visitors"`
	PrevVisitors    int             `json:"prev_visitors"`
	Daily           []DailyActivity `json:"daily"`
	Hourly          [24]int         `json:"hourly"`
	StatusBreakdown map[string]int  `json:"status_breakdown"`
	// TopTenants is only filled for the platform-wide view (nil tenant).
	TopTenants []TenantActivity `json:"top_tenants,omitempty"`
}

type TenantActivity struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Name     string    `json:"name"`
	User     int       `json:"user"`
	Bot      int       `json:"bot"`
}

type DailyActivity struct {
	Date string `json:"date"`
	User int    `json:"user"`
	Bot  int    `json:"bot"`
}

// ConversationStats aggregates the last `days` days and the equally long period before it, for trend comparison.
// A nil tenantID aggregates every tenant.
func (r Repository) ConversationStats(ctx context.Context, tenantID *uuid.UUID, days int) (ConversationStats, error) {
	var scope any
	if tenantID != nil {
		scope = tenantID.String()
	}
	out := ConversationStats{Days: days, Daily: []DailyActivity{}, StatusBreakdown: map[string]int{}}
	const tz = `'Asia/Jakarta'`
	err := r.DB.QueryRow(ctx, `
SELECT
  count(*) FILTER (WHERE m.sender_type='USER' AND m.created_at >= now()-make_interval(days=>$2)),
  count(*) FILTER (WHERE m.sender_type='BOT'  AND m.created_at >= now()-make_interval(days=>$2)),
  count(*) FILTER (WHERE m.sender_type='USER' AND m.created_at <  now()-make_interval(days=>$2)),
  count(*) FILTER (WHERE m.sender_type='BOT'  AND m.created_at <  now()-make_interval(days=>$2)),
  count(DISTINCT c.telegram_chat_id) FILTER (WHERE m.sender_type='USER' AND m.created_at >= now()-make_interval(days=>$2)),
  count(DISTINCT c.telegram_chat_id) FILTER (WHERE m.sender_type='USER' AND m.created_at <  now()-make_interval(days=>$2))
FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id
WHERE ($1::uuid IS NULL OR c.tenant_id=$1::uuid) AND m.created_at >= now()-make_interval(days=>$2*2)`, scope, days).
		Scan(&out.UserMessages, &out.BotReplies, &out.PrevUser, &out.PrevBot, &out.Visitors, &out.PrevVisitors)
	if err != nil {
		return out, err
	}
	// generate_series guarantees a row for every day, including days without traffic.
	rows, err := r.DB.Query(ctx, `
WITH d AS (SELECT generate_series((now() AT TIME ZONE `+tz+`)::date-($2-1), (now() AT TIME ZONE `+tz+`)::date, interval '1 day')::date AS day)
SELECT to_char(d.day,'YYYY-MM-DD'),
  count(m.id) FILTER (WHERE m.sender_type='USER'),
  count(m.id) FILTER (WHERE m.sender_type='BOT')
FROM d
LEFT JOIN (SELECT m.id, m.sender_type, (m.created_at AT TIME ZONE `+tz+`)::date AS day FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id WHERE ($1::uuid IS NULL OR c.tenant_id=$1::uuid)) m ON m.day=d.day
GROUP BY d.day ORDER BY d.day`, scope, days)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var item DailyActivity
		if err := rows.Scan(&item.Date, &item.User, &item.Bot); err != nil {
			rows.Close()
			return out, err
		}
		out.Daily = append(out.Daily, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	hours, err := r.DB.Query(ctx, `
SELECT extract(hour FROM m.created_at AT TIME ZONE `+tz+`)::int, count(*)
FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id
WHERE ($1::uuid IS NULL OR c.tenant_id=$1::uuid) AND m.sender_type='USER' AND m.created_at >= now()-make_interval(days=>$2)
GROUP BY 1`, scope, days)
	if err != nil {
		return out, err
	}
	for hours.Next() {
		var hour, n int
		if err := hours.Scan(&hour, &n); err != nil {
			hours.Close()
			return out, err
		}
		if hour >= 0 && hour < 24 {
			out.Hourly[hour] = n
		}
	}
	hours.Close()
	if err := hours.Err(); err != nil {
		return out, err
	}
	if tenantID == nil {
		top, err := r.DB.Query(ctx, `
SELECT t.id, t.name,
  count(*) FILTER (WHERE m.sender_type='USER'), count(*) FILTER (WHERE m.sender_type='BOT')
FROM telegram_messages m JOIN telegram_conversations c ON c.id=m.conversation_id JOIN tenants t ON t.id=c.tenant_id
WHERE m.created_at >= now()-make_interval(days=>$1)
GROUP BY t.id, t.name ORDER BY 3 DESC, t.name LIMIT 5`, days)
		if err != nil {
			return out, err
		}
		for top.Next() {
			var item TenantActivity
			if err := top.Scan(&item.TenantID, &item.Name, &item.User, &item.Bot); err != nil {
				top.Close()
				return out, err
			}
			out.TopTenants = append(out.TopTenants, item)
		}
		top.Close()
		if err := top.Err(); err != nil {
			return out, err
		}
	}
	status, err := r.DB.Query(ctx, `SELECT status, count(*) FROM telegram_conversations WHERE ($1::uuid IS NULL OR tenant_id=$1::uuid) GROUP BY status`, scope)
	if err != nil {
		return out, err
	}
	defer status.Close()
	for status.Next() {
		var key string
		var n int
		if err := status.Scan(&key, &n); err != nil {
			return out, err
		}
		out.StatusBreakdown[key] = n
	}
	return out, status.Err()
}

func (r Repository) GetWhatsAppChannel(ctx context.Context, tenantID uuid.UUID) (WhatsAppChannel, error) {
	var w WhatsAppChannel
	err := r.DB.QueryRow(ctx, `SELECT id,tenant_id,phone_number,status,access_mode,allowed_numbers,service_mode FROM whatsapp_channels WHERE tenant_id=$1`, tenantID).Scan(&w.ID, &w.TenantID, &w.PhoneNumber, &w.Status, &w.AccessMode, &w.AllowedNumbers, &w.ServiceMode)
	if w.AllowedNumbers == nil {
		w.AllowedNumbers = []string{}
	}
	return w, err
}

// UpsertWhatsAppChannel saves the linked number and access policy; the channel stays PENDING until the assistant starts.
func (r Repository) UpsertWhatsAppChannel(ctx context.Context, w WhatsAppChannel) error {
	_, err := r.DB.Exec(ctx, `INSERT INTO whatsapp_channels(tenant_id,phone_number,status,access_mode,allowed_numbers,service_mode) VALUES($1,$2,'PENDING',$3,$4,$5) ON CONFLICT (tenant_id) DO UPDATE SET phone_number=EXCLUDED.phone_number,access_mode=EXCLUDED.access_mode,allowed_numbers=EXCLUDED.allowed_numbers,service_mode=EXCLUDED.service_mode,status='PENDING',linked_at=now(),updated_at=now()`, w.TenantID, w.PhoneNumber, w.AccessMode, w.AllowedNumbers, w.ServiceMode)
	return err
}

// SetWhatsAppServiceMode changes only the service mode; the link and its status stay as they are.
func (r Repository) SetWhatsAppServiceMode(ctx context.Context, tenantID uuid.UUID, mode string) error {
	tag, err := r.DB.Exec(ctx, `UPDATE whatsapp_channels SET service_mode=$2, updated_at=now() WHERE tenant_id=$1`, tenantID, mode)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (r Repository) SetWhatsAppChannelStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	_, err := r.DB.Exec(ctx, `UPDATE whatsapp_channels SET status=$2, verified_at=CASE WHEN $2='ACTIVE' THEN now() ELSE verified_at END, updated_at=now() WHERE tenant_id=$1`, tenantID, status)
	return err
}

func (r Repository) DeleteWhatsAppChannel(ctx context.Context, tenantID uuid.UUID) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM whatsapp_channels WHERE tenant_id=$1`, tenantID)
	return err
}
