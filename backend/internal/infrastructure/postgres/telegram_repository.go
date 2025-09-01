package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/supercakecrumb/tg-analysis-tool/internal/domain"
)

type TelegramRepository struct {
	db *DB
}

func NewTelegramRepository(db *DB) *TelegramRepository {
	return &TelegramRepository{db: db}
}

func (r *TelegramRepository) SaveAuth(ctx context.Context, auth *domain.TelegramAuth) error {
	query := `
		INSERT INTO telegram_auth 
			(id, user_id, telegram_id, phone_number, phone_code_hash, session_data, is_authorized, created_at, updated_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			telegram_id = EXCLUDED.telegram_id,
			phone_number = EXCLUDED.phone_number,
			phone_code_hash = EXCLUDED.phone_code_hash,
			session_data = EXCLUDED.session_data,
			is_authorized = EXCLUDED.is_authorized,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		auth.ID,
		auth.UserID,
		auth.TelegramID,
		auth.PhoneNumber,
		auth.PhoneCodeHash,
		auth.SessionData,
		auth.IsAuthorized,
		auth.CreatedAt,
		auth.UpdatedAt,
	)

	return err
}

func (r *TelegramRepository) GetAuth(ctx context.Context, userID string) (*domain.TelegramAuth, error) {
	query := `
		SELECT id, user_id, telegram_id, phone_number, phone_code_hash, session_data, is_authorized, created_at, updated_at
		FROM telegram_auth
		WHERE user_id = $1
	`

	var auth domain.TelegramAuth
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&auth.ID,
		&auth.UserID,
		&auth.TelegramID,
		&auth.PhoneNumber,
		&auth.PhoneCodeHash,
		&auth.SessionData,
		&auth.IsAuthorized,
		&auth.CreatedAt,
		&auth.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &auth, nil
}

func (r *TelegramRepository) UpdateAuth(ctx context.Context, auth *domain.TelegramAuth) error {
	query := `
		UPDATE telegram_auth
		SET 
			telegram_id = $1,
			phone_number = $2,
			phone_code_hash = $3,
			session_data = $4,
			is_authorized = $5,
			updated_at = $6
		WHERE user_id = $7
	`

	result, err := r.db.Exec(ctx, query,
		auth.TelegramID,
		auth.PhoneNumber,
		auth.PhoneCodeHash,
		auth.SessionData,
		auth.IsAuthorized,
		auth.UpdatedAt,
		auth.UserID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("auth not found for user %s", auth.UserID)
	}

	return nil
}

func (r *TelegramRepository) DeleteAuth(ctx context.Context, userID string) error {
	query := `DELETE FROM telegram_auth WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *TelegramRepository) SaveChat(ctx context.Context, chat *domain.TelegramChat) error {
	query := `
		INSERT INTO telegram_chats (id, user_id, chat_id, title, type, username, member_count, last_sync_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, chat_id) DO UPDATE SET
			title = EXCLUDED.title,
			type = EXCLUDED.type,
			username = EXCLUDED.username,
			member_count = EXCLUDED.member_count,
			last_sync_at = EXCLUDED.last_sync_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		chat.ID,
		chat.UserID,
		chat.ChatID,
		chat.Title,
		chat.Type,
		chat.Username,
		chat.MemberCount,
		chat.LastSyncAt,
		chat.CreatedAt,
		chat.UpdatedAt,
	)

	return err
}

func (r *TelegramRepository) GetChats(ctx context.Context, userID string) ([]*domain.TelegramChat, error) {
	query := `
		SELECT id, user_id, chat_id, title, type, username, member_count, last_sync_at, created_at, updated_at
		FROM telegram_chats
		WHERE user_id = $1
		ORDER BY last_sync_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*domain.TelegramChat
	for rows.Next() {
		var chat domain.TelegramChat
		err := rows.Scan(
			&chat.ID,
			&chat.UserID,
			&chat.ChatID,
			&chat.Title,
			&chat.Type,
			&chat.Username,
			&chat.MemberCount,
			&chat.LastSyncAt,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}
	return chats, nil
}

func (r *TelegramRepository) GetChat(ctx context.Context, userID string, chatID int64) (*domain.TelegramChat, error) {
	query := `
		SELECT id, user_id, chat_id, title, type, username, member_count, last_sync_at, created_at, updated_at
		FROM telegram_chats
		WHERE user_id = $1 AND chat_id = $2
	`

	var chat domain.TelegramChat
	err := r.db.QueryRow(ctx, query, userID, chatID).Scan(
		&chat.ID,
		&chat.UserID,
		&chat.ChatID,
		&chat.Title,
		&chat.Type,
		&chat.Username,
		&chat.MemberCount,
		&chat.LastSyncAt,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *TelegramRepository) SaveMessages(ctx context.Context, messages []*domain.TelegramMessage) error {
	if len(messages) == 0 {
		return nil
	}

	return r.db.Transaction(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			CREATE TEMP TABLE temp_messages (
				id BIGINT,
				user_id TEXT,
				chat_id BIGINT,
				message_id INTEGER,
				from_id BIGINT,
				text TEXT,
				date TIMESTAMP,
				reply_to_id INTEGER,
				forward_id BIGINT,
				media_type TEXT,
				has_emoji BOOLEAN,
				word_count INTEGER,
				is_outgoing BOOLEAN,
				created_at TIMESTAMP
			) ON COMMIT DROP
		`)
		if err != nil {
			return err
		}

		rows := make([][]interface{}, len(messages))
		for i, msg := range messages {
			rows[i] = []interface{}{
				msg.ID,
				msg.UserID,
				msg.ChatID,
				msg.MessageID,
				msg.FromID,
				msg.Text,
				msg.Date,
				msg.ReplyToID,
				msg.ForwardID,
				msg.MediaType,
				msg.HasEmoji,
				msg.WordCount,
				msg.IsOutgoing,
				msg.CreatedAt,
			}
		}

		_, err = tx.CopyFrom(ctx,
			pgx.Identifier{"temp_messages"},
			[]string{"id", "user_id", "chat_id", "message_id", "from_id", "text", "date",
				"reply_to_id", "forward_id", "media_type", "has_emoji", "word_count",
				"is_outgoing", "created_at"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO telegram_messages 
			SELECT * FROM temp_messages
			ON CONFLICT (user_id, chat_id, message_id) DO NOTHING
		`)

		return err
	})
}

func (r *TelegramRepository) GetMessages(ctx context.Context, userID string, chatID int64, limit, offset int) ([]*domain.TelegramMessage, error) {
	query := `
        SELECT id, user_id, chat_id, message_id, from_id, text, date,
               reply_to_id, forward_id, media_type, has_emoji, word_count,
               is_outgoing, created_at
        FROM telegram_messages
        WHERE user_id = $1 AND chat_id = $2
        ORDER BY date DESC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.Query(ctx, query, userID, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.TelegramMessage
	for rows.Next() {
		var msg domain.TelegramMessage
		err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.ChatID,
			&msg.MessageID,
			&msg.FromID,
			&msg.Text,
			&msg.Date,
			&msg.ReplyToID,
			&msg.ForwardID,
			&msg.MediaType,
			&msg.HasEmoji,
			&msg.WordCount,
			&msg.IsOutgoing,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *TelegramRepository) GetMessageCount(ctx context.Context, userID string, chatID int64) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM telegram_messages 
		WHERE user_id = $1 AND chat_id = $2
	`

	var count int64
	err := r.db.QueryRow(ctx, query, userID, chatID).Scan(&count)
	return count, err
}
