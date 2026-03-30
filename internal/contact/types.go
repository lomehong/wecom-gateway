package contact

import "wecom-gateway/internal/wecom"

// GetUserListRequest represents a request to list contact users
type GetUserListRequest struct {
	DepartmentID int `form:"department_id"`
}

// GetUserListResponse represents the response for listing contact users
type GetUserListResponse struct {
	Users []*wecom.ContactUser `json:"users"`
	Count int                  `json:"count"`
}

// SearchUserRequest represents a request to search contact users
type SearchUserRequest struct {
	Query string `form:"query" binding:"required"`
}

// SearchUserResponse represents the response for searching contact users
type SearchUserResponse struct {
	Users []*wecom.ContactUser `json:"users"`
	Count int                  `json:"count"`
}
