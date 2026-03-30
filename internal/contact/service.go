package contact

import (
	"context"
	"fmt"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Service handles contact business logic
type Service struct {
	wecomClient wecom.Client
}

// NewService creates a new contact service
func NewService(wecomClient wecom.Client) *Service {
	return &Service{
		wecomClient: wecomClient,
	}
}

// GetUserList retrieves the list of contact users for a department
func (s *Service) GetUserList(ctx context.Context, authCtx *auth.AuthContext, req *GetUserListRequest) (*GetUserListResponse, error) {
	// Set default department ID to root (1)
	departmentID := req.DepartmentID
	if departmentID == 0 {
		departmentID = 1
	}

	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	users, err := s.wecomClient.GetUserList(ctx, corpName, appName, departmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user list: %w", err)
	}

	return &GetUserListResponse{
		Users: users,
		Count: len(users),
	}, nil
}

// SearchUser searches for contact users by name, alias, or user ID
func (s *Service) SearchUser(ctx context.Context, authCtx *auth.AuthContext, req *SearchUserRequest) (*SearchUserResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Determine corp and app from auth context
	corpName := authCtx.CorpName
	appName := authCtx.AppName

	users, err := s.wecomClient.SearchUser(ctx, corpName, appName, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return &SearchUserResponse{
		Users: users,
		Count: len(users),
	}, nil
}
