package todo

import (
	"context"
	"fmt"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/wecom"
)

// Service handles todo business logic
type Service struct {
	wecomClient wecom.Client
}

// NewService creates a new todo service
func NewService(wecomClient wecom.Client) *Service {
	return &Service{
		wecomClient: wecomClient,
	}
}

// GetTodoList retrieves the list of todos
func (s *Service) GetTodoList(ctx context.Context, authCtx *auth.AuthContext, req *GetTodoListRequest) (*wecom.TodoListResult, error) {
	// Set defaults
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	opts := &wecom.TodoListOptions{
		CreateBeginTime: req.CreateBeginTime,
		CreateEndTime:   req.CreateEndTime,
		RemindBeginTime: req.RemindBeginTime,
		RemindEndTime:   req.RemindEndTime,
		Limit:           limit,
		Cursor:          req.Cursor,
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	result, err := s.wecomClient.GetTodoList(ctx, corpName, appName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo list: %w", err)
	}

	return result, nil
}

// GetTodoDetail retrieves the detail of a specific todo
func (s *Service) GetTodoDetail(ctx context.Context, authCtx *auth.AuthContext, todoID string) ([]*wecom.TodoDetail, error) {
	if todoID == "" {
		return nil, fmt.Errorf("todo id is required")
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	details, err := s.wecomClient.GetTodoDetail(ctx, corpName, appName, []string{todoID})
	if err != nil {
		return nil, fmt.Errorf("failed to get todo detail: %w", err)
	}

	return details, nil
}

// CreateTodo creates a new todo
func (s *Service) CreateTodo(ctx context.Context, authCtx *auth.AuthContext, req *CreateTodoRequest) (string, error) {
	if req.Content == "" {
		return "", fmt.Errorf("content is required")
	}

	params := &wecom.CreateTodoParams{
		Title:      req.Title,
		Content:    req.Content,
		Assignees:  req.Assignees,
		RemindTime: req.RemindTime,
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	todoID, err := s.wecomClient.CreateTodo(ctx, corpName, appName, params)
	if err != nil {
		return "", fmt.Errorf("failed to create todo: %w", err)
	}

	return todoID, nil
}

// UpdateTodo updates an existing todo
func (s *Service) UpdateTodo(ctx context.Context, authCtx *auth.AuthContext, todoID string, req *UpdateTodoRequest) error {
	if todoID == "" {
		return fmt.Errorf("todo id is required")
	}

	params := &wecom.UpdateTodoParams{
		Title:      req.Title,
		Content:    req.Content,
		Status:     req.Status,
		RemindTime: req.RemindTime,
		Assignees:  req.Assignees,
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	err := s.wecomClient.UpdateTodo(ctx, corpName, appName, todoID, params)
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	return nil
}

// DeleteTodo deletes a todo
func (s *Service) DeleteTodo(ctx context.Context, authCtx *auth.AuthContext, todoID string) error {
	if todoID == "" {
		return fmt.Errorf("todo id is required")
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	err := s.wecomClient.DeleteTodo(ctx, corpName, appName, todoID)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// ChangeUserStatus changes the user processing status of a todo
func (s *Service) ChangeUserStatus(ctx context.Context, authCtx *auth.AuthContext, todoID string, status int) error {
	if todoID == "" {
		return fmt.Errorf("todo id is required")
	}

	// Validate status value
	if status != 0 && status != 1 {
		return fmt.Errorf("invalid status value, must be 0 (completed) or 1 (in progress)")
	}

	corpName := authCtx.CorpName
	appName := authCtx.AppName

	err := s.wecomClient.ChangeTodoUserStatus(ctx, corpName, appName, todoID, status)
	if err != nil {
		return fmt.Errorf("failed to change todo user status: %w", err)
	}

	return nil
}
