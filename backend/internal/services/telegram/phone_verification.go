package telegram

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/supercakecrumb/tg-analysis-tool/internal/config"
	"github.com/supercakecrumb/tg-analysis-tool/internal/domain"
)

type PhoneVerifier struct {
	config *config.Config
	repo   domain.TelegramRepository
}

func NewPhoneVerifier(cfg *config.Config, repo domain.TelegramRepository) *PhoneVerifier {
	return &PhoneVerifier{
		config: cfg,
		repo:   repo,
	}
}

func (p *PhoneVerifier) SendCode(ctx context.Context, userID, phoneNumber string) (*domain.PhoneCodeAuth, error) {
	opts := telegram.Options{
		SessionStorage: &session.StorageMemory{},
	}
	client := telegram.NewClient(p.config.TelegramAPIID, p.config.TelegramAPIHash, opts)

	var phoneCodeHash string
	result, err := client.Auth().SendCode(ctx, phoneNumber, auth.SendCodeOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to send verification code: %w", err)
	}
	if sentCode, ok := result.(*tg.AuthSentCode); ok {
		phoneCodeHash = sentCode.PhoneCodeHash
	} else {
		return nil, fmt.Errorf("unexpected sent code type: %T", result)
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

	if err := p.repo.SaveAuth(ctx, authData); err != nil {
		return nil, fmt.Errorf("failed to save auth: %w", err)
	}

	return &domain.PhoneCodeAuth{
		PhoneNumber: phoneNumber,
		CodeHash:    phoneCodeHash,
	}, nil
}

func (p *PhoneVerifier) VerifyCode(ctx context.Context, userID, phoneNumber, code, codeHash string) error {
	authData, err := p.repo.GetAuth(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get auth data: %w", err)
	}
	if authData == nil {
		return errors.New("auth not found")
	}

	opts := telegram.Options{
		SessionStorage: &session.StorageMemory{},
	}
	client := telegram.NewClient(p.config.TelegramAPIID, p.config.TelegramAPIHash, opts)

	authResult, err := client.Auth().SignIn(ctx, phoneNumber, code, authData.PhoneCodeHash)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	authData.IsAuthorized = true
	// authResult is already *tg.AuthAuthorization
	if user, ok := authResult.User.(*tg.User); ok {
		authData.TelegramID = user.ID
	} else {
		return fmt.Errorf("unexpected user type: %T", authResult.User)
	}
	authData.UpdatedAt = time.Now()

	return p.repo.UpdateAuth(ctx, authData)
}
