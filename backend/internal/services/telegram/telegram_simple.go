package telegram

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/domain"
)

// SimpleService implements a simplified Telegram service for compilation
type SimpleService struct {
	config     *config.Config
	repo       domain.TelegramRepository
	encryptKey []byte
}

// NewSimpleService creates a new simplified Telegram service
func NewSimpleService(cfg *config.Config, repo domain.TelegramRepository) domain.TelegramService {
	return &SimpleService{
		config:     cfg,
		repo:       repo,
		encryptKey: []byte(cfg.SessionEncKey),
	}
}

// GenerateQRCode is temporarily disabled due to dependency issues
func (s *SimpleService) GenerateQRCode(ctx context.Context, userID string) (*domain.QRCodeAuth, error) {
	return nil, fmt.Errorf("QR code authentication is temporarily unavailable")
}

// SendPhoneCode sends a verification code to the phone number
func (s *SimpleService) SendPhoneCode(ctx context.Context, userID, phoneNumber string) (*domain.PhoneCodeAuth, error) {
	opts := telegram.Options{
		SessionStorage: &session.StorageMemory{},
	}
	client := telegram.NewClient(s.config.TelegramAPIID, s.config.TelegramAPIHash, opts)

	var phoneCodeHash string
	sentCode, err := client.Auth().SendCode(ctx, phoneNumber, auth.SendCodeOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}
	if s, ok := sentCode.(*tg.AuthSentCode); ok {
		phoneCodeHash = s.PhoneCodeHash
	} else {
		return nil, fmt.Errorf("unexpected sent code type: %T", sentCode)
	}

	authData := &domain.TelegramAuth{
		ID:            fmt.Sprintf("%d", time.Now().UnixNano()),
		UserID:        userID,
		PhoneNumber:   phoneNumber,
		PhoneCodeHash: phoneCodeHash,
		IsAuthorized:  false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	sessionData, err := s.encryptSession([]byte("pending-verification"))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt session: %w", err)
	}
	authData.SessionData = sessionData

	if err := s.repo.SaveAuth(ctx, authData); err != nil {
		return nil, fmt.Errorf("failed to save auth: %w", err)
	}

	return &domain.PhoneCodeAuth{
		PhoneNumber: phoneNumber,
		CodeHash:    phoneCodeHash,
	}, nil
}

// VerifyPhoneCode verifies the phone code
func (s *SimpleService) VerifyPhoneCode(ctx context.Context, userID, phoneNumber, code, codeHash string) error {
	authData, err := s.repo.GetAuth(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get auth data: %w", err)
	}
	if authData == nil {
		return errors.New("auth not found")
	}

	opts := telegram.Options{
		SessionStorage: &session.StorageMemory{},
	}
	client := telegram.NewClient(s.config.TelegramAPIID, s.config.TelegramAPIHash, opts)

	authUser, err := client.Auth().SignIn(ctx, phoneNumber, code, authData.PhoneCodeHash)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	authData.IsAuthorized = true
	if user, ok := authUser.User.(*tg.User); ok {
		authData.TelegramID = user.ID
	} else {
		return fmt.Errorf("unexpected user type: %T", authUser.User)
	}
	authData.UpdatedAt = time.Now()

	sessionData, err := s.encryptSession([]byte("TODO: implement proper session serialization"))
	if err != nil {
		return fmt.Errorf("failed to encrypt session: %w", err)
	}
	authData.SessionData = sessionData

	return s.repo.UpdateAuth(ctx, authData)
}

// VerifyPassword verifies the 2FA password
func (s *SimpleService) VerifyPassword(ctx context.Context, userID, password string) error {
	// Simplified implementation
	authData, err := s.repo.GetAuth(ctx, userID)
	if err != nil {
		return err
	}
	if authData == nil {
		return errors.New("auth not found")
	}

	// In real implementation, this would verify with Telegram
	authData.IsAuthorized = true
	authData.UpdatedAt = time.Now()
	return s.repo.UpdateAuth(ctx, authData)
}

// GetAuthStatus returns the current authentication status
func (s *SimpleService) GetAuthStatus(ctx context.Context, userID string) (*domain.TelegramAuth, error) {
	return s.repo.GetAuth(ctx, userID)
}

// Logout logs out the user from Telegram
func (s *SimpleService) Logout(ctx context.Context, userID string) error {
	return s.repo.DeleteAuth(ctx, userID)
}

// SyncChats synchronizes all chats for a user
func (s *SimpleService) SyncChats(ctx context.Context, userID string) error {
	// Simplified implementation - create some mock chats
	chats := []*domain.TelegramChat{
		{
			ID:          time.Now().UnixNano(),
			UserID:      userID,
			ChatID:      1001,
			Title:       "Test Chat 1",
			Type:        "group",
			MemberCount: 10,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          time.Now().UnixNano() + 1,
			UserID:      userID,
			ChatID:      1002,
			Title:       "Test Channel",
			Type:        "channel",
			MemberCount: 100,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, chat := range chats {
		if err := s.repo.SaveChat(ctx, chat); err != nil {
			return err
		}
	}

	return nil
}

// SyncMessages synchronizes messages for a specific chat
func (s *SimpleService) SyncMessages(ctx context.Context, userID string, chatID int64, limit int) error {
	// Simplified implementation - create some mock messages
	messages := make([]*domain.TelegramMessage, 0, limit)

	for i := 0; i < limit && i < 10; i++ {
		msg := &domain.TelegramMessage{
			ID:         time.Now().UnixNano() + int64(i),
			UserID:     userID,
			ChatID:     chatID,
			MessageID:  i + 1,
			FromID:     12345,
			Text:       fmt.Sprintf("Test message %d", i+1),
			Date:       time.Now().Add(-time.Duration(i) * time.Hour),
			IsOutgoing: i%2 == 0,
			WordCount:  3,
			HasEmoji:   false,
			CreatedAt:  time.Now(),
		}
		messages = append(messages, msg)
	}

	return s.repo.SaveMessages(ctx, messages)
}

// GetSyncProgress returns the current sync progress
func (s *SimpleService) GetSyncProgress(ctx context.Context, userID string) (*domain.SyncProgress, error) {
	// Simplified implementation
	return &domain.SyncProgress{
		UserID:         userID,
		Status:         "idle",
		TotalChats:     2,
		SyncedChats:    2,
		TotalMessages:  20,
		SyncedMessages: 20,
	}, nil
}

// Helper methods

func (s *SimpleService) encryptSession(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (s *SimpleService) decryptSession(encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
