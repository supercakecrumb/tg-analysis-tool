package domain

import (
	"context"
	"time"
)

// TelegramAuth represents the authentication state for a Telegram session
type TelegramAuth struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	TelegramID    int64     `json:"telegram_id" db:"telegram_id"`
	PhoneNumber   string    `json:"phone_number" db:"phone_number"`
	PhoneCodeHash string    `json:"-" db:"phone_code_hash"` // Temporary hash for phone verification
	SessionData   []byte    `json:"-" db:"session_data"`    // Encrypted session data
	IsAuthorized  bool      `json:"is_authorized" db:"is_authorized"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	PhotoURL    string `json:"photo_url"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID          int64     `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	ChatID      int64     `json:"chat_id" db:"chat_id"`
	Title       string    `json:"title" db:"title"`
	Type        string    `json:"type" db:"type"` // private, group, supergroup, channel
	Username    string    `json:"username" db:"username"`
	MemberCount int       `json:"member_count" db:"member_count"`
	LastSyncAt  time.Time `json:"last_sync_at" db:"last_sync_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	ID         int64     `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	ChatID     int64     `json:"chat_id" db:"chat_id"`
	MessageID  int       `json:"message_id" db:"message_id"`
	FromID     int64     `json:"from_id" db:"from_id"`
	Text       string    `json:"text" db:"text"`
	Date       time.Time `json:"date" db:"date"`
	ReplyToID  *int      `json:"reply_to_id" db:"reply_to_id"`
	ForwardID  *int64    `json:"forward_id" db:"forward_id"`
	MediaType  string    `json:"media_type" db:"media_type"`
	HasEmoji   bool      `json:"has_emoji" db:"has_emoji"`
	WordCount  int       `json:"word_count" db:"word_count"`
	IsOutgoing bool      `json:"is_outgoing" db:"is_outgoing"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// QRCodeAuth represents QR code authentication data
type QRCodeAuth struct {
	Token     string    `json:"token"`
	QRURL     string    `json:"qr_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PhoneCodeAuth represents phone code authentication data
type PhoneCodeAuth struct {
	PhoneNumber string `json:"phone_number"`
	CodeHash    string `json:"code_hash"`
}

// TelegramService defines the interface for Telegram operations
type TelegramService interface {
	// Authentication methods
	GenerateQRCode(ctx context.Context, userID string) (*QRCodeAuth, error)
	SendPhoneCode(ctx context.Context, userID, phoneNumber string) (*PhoneCodeAuth, error)
	VerifyPhoneCode(ctx context.Context, userID, phoneNumber, code, codeHash string) error
	VerifyPassword(ctx context.Context, userID, password string) error
	GetAuthStatus(ctx context.Context, userID string) (*TelegramAuth, error)
	Logout(ctx context.Context, userID string) error

	// Data sync methods
	SyncChats(ctx context.Context, userID string) error
	SyncMessages(ctx context.Context, userID string, chatID int64, limit int) error
	GetSyncProgress(ctx context.Context, userID string) (*SyncProgress, error)
}

// TelegramRepository defines the interface for Telegram data persistence
type TelegramRepository interface {
	// Auth methods
	SaveAuth(ctx context.Context, auth *TelegramAuth) error
	GetAuth(ctx context.Context, userID string) (*TelegramAuth, error)
	UpdateAuth(ctx context.Context, auth *TelegramAuth) error
	DeleteAuth(ctx context.Context, userID string) error

	// Chat methods
	SaveChat(ctx context.Context, chat *TelegramChat) error
	GetChats(ctx context.Context, userID string) ([]*TelegramChat, error)
	GetChat(ctx context.Context, userID string, chatID int64) (*TelegramChat, error)

	// Message methods
	SaveMessages(ctx context.Context, messages []*TelegramMessage) error
	GetMessages(ctx context.Context, userID string, chatID int64, limit, offset int) ([]*TelegramMessage, error)
	GetMessageCount(ctx context.Context, userID string, chatID int64) (int64, error)
}

// SyncProgress represents the progress of data synchronization
type SyncProgress struct {
	UserID         string     `json:"user_id"`
	Status         string     `json:"status"` // idle, syncing, completed, failed
	TotalChats     int        `json:"total_chats"`
	SyncedChats    int        `json:"synced_chats"`
	TotalMessages  int64      `json:"total_messages"`
	SyncedMessages int64      `json:"synced_messages"`
	CurrentChat    string     `json:"current_chat"`
	LastError      string     `json:"last_error,omitempty"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}
