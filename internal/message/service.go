package message

import (
	"context"
	"fmt"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Service handles message business logic
type Service struct {
	wecomClient wecom.Client
}

// NewService creates a new message service
func NewService(wecomClient wecom.Client) *Service {
	return &Service{wecomClient: wecomClient}
}

// SendText sends a text message
func (s *Service) SendText(ctx context.Context, authCtx *auth.AuthContext, receiverType string, receiverIDs []string, content string, safe bool) (*wecom.SendResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.MessageParams{
		ReceiverType: receiverType,
		ReceiverIDs:  receiverIDs,
		Content:      content,
		Safe:         safe,
	}

	result, err := s.wecomClient.SendText(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send text message: %w", err)
	}

	return result, nil
}

// SendMarkdown sends a markdown message
func (s *Service) SendMarkdown(ctx context.Context, authCtx *auth.AuthContext, receiverType string, receiverIDs []string, content string) (*wecom.SendResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.MessageParams{
		ReceiverType: receiverType,
		ReceiverIDs:  receiverIDs,
		Content:      content,
	}

	result, err := s.wecomClient.SendMarkdown(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send markdown message: %w", err)
	}

	return result, nil
}

// SendImage sends an image message
func (s *Service) SendImage(ctx context.Context, authCtx *auth.AuthContext, receiverType string, receiverIDs []string, mediaID string, imageURL string) (*wecom.SendResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.ImageMessageParams{
		ReceiverType: receiverType,
		ReceiverIDs:  receiverIDs,
		MediaID:      mediaID,
		ImageURL:     imageURL,
	}

	result, err := s.wecomClient.SendImage(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send image message: %w", err)
	}

	return result, nil
}

// SendFile sends a file message
func (s *Service) SendFile(ctx context.Context, authCtx *auth.AuthContext, receiverType string, receiverIDs []string, mediaID string, fileURL string) (*wecom.SendResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.FileMessageParams{
		ReceiverType: receiverType,
		ReceiverIDs:  receiverIDs,
		MediaID:      mediaID,
		FileURL:      fileURL,
	}

	result, err := s.wecomClient.SendFile(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send file message: %w", err)
	}

	return result, nil
}

// SendCard sends a card message
func (s *Service) SendCard(ctx context.Context, authCtx *auth.AuthContext, receiverType string, receiverIDs []string, cardContent map[string]interface{}) (*wecom.SendResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	params := &wecom.CardMessageParams{
		ReceiverType: receiverType,
		ReceiverIDs:  receiverIDs,
		CardContent: cardContent,
	}

	result, err := s.wecomClient.SendCard(ctx, corpName, appName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to send card message: %w", err)
	}

	return result, nil
}

// --- Message Pull Service Methods (Phase 3.1) ---

// GetChatList gets the list of chat conversations
func (s *Service) GetChatList(ctx context.Context, authCtx *auth.AuthContext, beginTime, endTime int64) (*wecom.ChatListResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	result, err := s.wecomClient.GetChatList(ctx, corpName, appName, beginTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat list: %w", err)
	}

	return result, nil
}

// GetChatMessages pulls messages from a chat conversation
func (s *Service) GetChatMessages(ctx context.Context, authCtx *auth.AuthContext, chatType int, chatID string, beginTime, endTime int64) (*wecom.ChatMessagesResult, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	result, err := s.wecomClient.GetChatMessages(ctx, corpName, appName, chatType, chatID, beginTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	return result, nil
}

// DownloadMedia downloads a media file
func (s *Service) DownloadMedia(ctx context.Context, authCtx *auth.AuthContext, mediaID string) ([]byte, string, error) {
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	data, filename, err := s.wecomClient.DownloadMedia(ctx, corpName, appName, mediaID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download media: %w", err)
	}

	return data, filename, nil
}
