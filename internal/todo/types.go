package todo

import "time"

// GetTodoListRequest represents a request to list todos
type GetTodoListRequest struct {
	CreateBeginTime *time.Time `form:"create_begin_time"`
	CreateEndTime   *time.Time `form:"create_end_time"`
	RemindBeginTime *time.Time `form:"remind_begin_time"`
	RemindEndTime   *time.Time `form:"remind_end_time"`
	Limit           int        `form:"limit"`
	Cursor          string     `form:"cursor"`
}

// CreateTodoRequest represents a request to create a todo
type CreateTodoRequest struct {
	Title      string     `json:"title,omitempty"`
	Content    string     `json:"content" binding:"required"`
	Assignees  []string   `json:"assignees,omitempty"`
	RemindTime *time.Time `json:"remind_time,omitempty"`
}

// UpdateTodoRequest represents a request to update a todo
type UpdateTodoRequest struct {
	Title      *string     `json:"title,omitempty"`
	Content    *string     `json:"content,omitempty"`
	Status     *int        `json:"status,omitempty"`
	RemindTime *time.Time  `json:"remind_time,omitempty"`
	Assignees  []string    `json:"assignees,omitempty"`
}

// ChangeUserStatusRequest represents a request to change user status on a todo
type ChangeUserStatusRequest struct {
	Status int `json:"status" binding:"required"`
}

// GetTodoDetailRequest represents a request to get todo detail
type GetTodoDetailRequest struct {
	ID string `uri:"id" binding:"required"`
}
