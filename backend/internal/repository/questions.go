package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"ponpes-saas/internal/links"
	"ponpes-saas/internal/questions"
)

// ErrQuestionLimit means the tenant already has the maximum number of answers handed to the assistant.
var ErrQuestionLimit = errors.New("answered question limit reached")

// TenantQuestion is a question the assistant could not answer, and what the admin did about it.
type TenantQuestion struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Question     string    `json:"question"`
	AskCount     int       `json:"ask_count"`
	Status       string    `json:"status"`
	FAQQuestion  string    `json:"faq_question"`
	Answer       string    `json:"answer"`
	FirstAskedAt time.Time `json:"first_asked_at"`
	LastAskedAt  time.Time `json:"last_asked_at"`
}

const tenantQuestionColumns = `id, tenant_id, question, ask_count, status, faq_question, answer, first_asked_at, last_asked_at`

func scanTenantQuestion(row pgx.Row) (TenantQuestion, error) {
	var q TenantQuestion
	err := row.Scan(&q.ID, &q.TenantID, &q.Question, &q.AskCount, &q.Status, &q.FAQQuestion, &q.Answer, &q.FirstAskedAt, &q.LastAskedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return q, ErrNotFound
	}
	return q, err
}

// LastUserMessage is the newest thing the visitor wrote in a conversation ("" when there is none).
func (r Repository) LastUserMessage(ctx context.Context, conversationID uuid.UUID) string {
	var text string
	if err := r.DB.QueryRow(ctx, `SELECT message_text FROM telegram_messages WHERE conversation_id=$1 AND sender_type='USER' ORDER BY created_at DESC LIMIT 1`, conversationID).Scan(&text); err != nil {
		return ""
	}
	return text
}

// RecordUnanswered files a question the assistant could not answer. The same question again only raises its count;
// a resolved one reopens (the fix did not work), while answered and ignored ones keep their status.
func (r Repository) RecordUnanswered(ctx context.Context, tenantID uuid.UUID, question string) error {
	question = links.Clean(question, 500)
	key := questions.Key(question)
	if question == "" || key == "" {
		return nil
	}
	if _, err := r.DB.Exec(ctx, `INSERT INTO tenant_questions(tenant_id,question,question_key) VALUES($1,$2,$3)
		ON CONFLICT (tenant_id,question_key) DO UPDATE SET ask_count=tenant_questions.ask_count+1, last_asked_at=now(), updated_at=now(),
		  question=EXCLUDED.question,
		  status=CASE WHEN tenant_questions.status IN ('ANSWERED','IGNORED') THEN tenant_questions.status ELSE 'OPEN' END`, tenantID, question, key); err != nil {
		return err
	}
	_, err := r.DB.Exec(ctx, `DELETE FROM tenant_questions WHERE tenant_id=$1 AND status='OPEN'
		AND id NOT IN (SELECT id FROM tenant_questions WHERE tenant_id=$1 AND status='OPEN' ORDER BY last_asked_at DESC LIMIT $2)`, tenantID, questions.MaxOpen)
	return err
}

// ListTenantQuestions returns the questions in one status, most asked first.
func (r Repository) ListTenantQuestions(ctx context.Context, tenantID uuid.UUID, status string) ([]TenantQuestion, error) {
	rows, err := r.DB.Query(ctx, `SELECT `+tenantQuestionColumns+` FROM tenant_questions WHERE tenant_id=$1 AND status=$2 ORDER BY ask_count DESC, last_asked_at DESC LIMIT 200`, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TenantQuestion{}
	for rows.Next() {
		q, err := scanTenantQuestion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// CountTenantQuestions returns how many questions are in each status.
func (r Repository) CountTenantQuestions(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	rows, err := r.DB.Query(ctx, `SELECT status, count(*) FROM tenant_questions WHERE tenant_id=$1 GROUP BY status`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{questions.StatusOpen: 0, questions.StatusAnswered: 0, questions.StatusResolved: 0, questions.StatusIgnored: 0}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			return nil, err
		}
		out[status] = n
	}
	return out, rows.Err()
}

func (r Repository) GetTenantQuestion(ctx context.Context, tenantID, id uuid.UUID) (TenantQuestion, error) {
	return scanTenantQuestion(r.DB.QueryRow(ctx, `SELECT `+tenantQuestionColumns+` FROM tenant_questions WHERE id=$1 AND tenant_id=$2`, id, tenantID))
}

// UpdateTenantQuestion saves a new status and, for ANSWERED, the answer. At most MaxAnswered answers are kept per tenant.
func (r Repository) UpdateTenantQuestion(ctx context.Context, q TenantQuestion) (TenantQuestion, error) {
	tag, err := r.DB.Exec(ctx, `UPDATE tenant_questions SET status=$3, faq_question=$4, answer=$5, updated_at=now()
		WHERE id=$1 AND tenant_id=$2 AND ($3<>'ANSWERED' OR status='ANSWERED' OR (SELECT count(*) FROM tenant_questions WHERE tenant_id=$2 AND status='ANSWERED') < $6)`,
		q.ID, q.TenantID, q.Status, q.FAQQuestion, q.Answer, questions.MaxAnswered)
	if err != nil {
		return q, err
	}
	if tag.RowsAffected() == 0 {
		if _, err := r.GetTenantQuestion(ctx, q.TenantID, q.ID); err != nil {
			return q, err
		}
		return q, ErrQuestionLimit
	}
	return r.GetTenantQuestion(ctx, q.TenantID, q.ID)
}

func (r Repository) DeleteTenantQuestion(ctx context.Context, tenantID, id uuid.UUID) error {
	tag, err := r.DB.Exec(ctx, `DELETE FROM tenant_questions WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

// AnsweredQuestionsByProfileKey lists the admin's answers for the assistant, oldest first so the order stays stable.
func (r Repository) AnsweredQuestionsByProfileKey(ctx context.Context, profileKey string) []questions.Answered {
	rows, err := r.DB.Query(ctx, `SELECT coalesce(nullif(q.faq_question,''),q.question), q.answer FROM tenant_questions q JOIN hermes_profiles h ON h.tenant_id=q.tenant_id
		WHERE h.profile_key=$1 AND q.status='ANSWERED' AND q.answer<>'' ORDER BY q.updated_at, q.id`, profileKey)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []questions.Answered
	for rows.Next() {
		var a questions.Answered
		if err := rows.Scan(&a.Question, &a.Answer); err != nil {
			return nil
		}
		out = append(out, a)
	}
	return out
}
