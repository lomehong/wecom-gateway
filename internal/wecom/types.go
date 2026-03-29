package wecom

import "time"

// Schedule represents a WeChat Work schedule
type Schedule struct {
	ScheduleID      string    `json:"schedule_id"`
	Organizer       string    `json:"organizer"`
	Summary         string    `json:"summary"`
	Description     string    `json:"description,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Attendees       []string  `json:"attendees,omitempty"`
	Location        string    `json:"location,omitempty"`
	RemindBeforeMin int       `json:"remind_before_minutes,omitempty"`
	CalID           string    `json:"cal_id"`
}

// ScheduleParams represents parameters for creating/updating a schedule
type ScheduleParams struct {
	Organizer       string    `json:"organizer"`
	Summary         string    `json:"summary"`
	Description     string    `json:"description,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Attendees       []string  `json:"attendees,omitempty"`
	Location        string    `json:"location,omitempty"`
	RemindBeforeMin int       `json:"remind_before_minutes,omitempty"`
}

// MeetingRoom represents a WeChat Work meeting room
type MeetingRoom struct {
	MBookingID   string            `json:"mbooking_id"`
	Name         string            `json:"name"`
	Capacity     int               `json:"capacity"`
	City         string            `json:"city"`
	Building     string            `json:"building"`
	Floor        string            `json:"floor"`
	Equipment    []string          `json:"equipment,omitempty"`
	Description  string            `json:"description,omitempty"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

// TimeSlot represents a time slot for room availability
type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// BookingResult represents the result of booking a meeting room
type BookingResult struct {
	BookingID   string    `json:"booking_id"`
	ScheduleID  string    `json:"schedule_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	MeetingRoom *MeetingRoom `json:"meeting_room"`
}

// BookingParams represents parameters for booking a meeting room
type BookingParams struct {
	MeetingRoomID string   `json:"meetingroom_id"`
	Subject       string   `json:"subject"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Booker        string   `json:"booker"`
	Attendees     []string `json:"attendees,omitempty"`
}

// RoomQueryOptions represents options for querying meeting rooms
type RoomQueryOptions struct {
	City     string
	Building string
	Floor    string
	Capacity int
	Equipment []string
	Cursor   string
	Limit    int
}

// MessageParams represents common parameters for sending messages
type MessageParams struct {
	ReceiverType string   `json:"receiver_type"` // user, department, tag
	ReceiverIDs  []string `json:"receiver_ids"`
	Content      string   `json:"content"`
	Safe         bool     `json:"safe"`
}

// ImageMessageParams represents parameters for sending image messages
type ImageMessageParams struct {
	ReceiverType string   `json:"receiver_type"`
	ReceiverIDs  []string `json:"receiver_ids"`
	ImageURL     string   `json:"image_url,omitempty"`
	MediaID      string   `json:"media_id,omitempty"`
}

// FileMessageParams represents parameters for sending file messages
type FileMessageParams struct {
	ReceiverType string   `json:"receiver_type"`
	ReceiverIDs  []string `json:"receiver_ids"`
	FileURL      string   `json:"file_url,omitempty"`
	MediaID      string   `json:"media_id,omitempty"`
}

// CardMessageParams represents parameters for sending card messages
type CardMessageParams struct {
	ReceiverType string                 `json:"receiver_type"`
	ReceiverIDs  []string               `json:"receiver_ids"`
	CardContent  map[string]interface{} `json:"card_content"`
}

// SendResult represents the result of sending a message
type SendResult struct {
	InvalidUserIDs    []string `json:"invalid_user_ids"`
	InvalidPartyIDs   []string `json:"invalid_party_ids"`
	InvalidTagIDs     []string `json:"invalid_tag_ids"`
	UnquotedUserIDs   []string `json:"unquoted_user_ids"`
	FailedUserIDs     []string `json:"failed_user_ids"`
	FailedPartyIDs    []string `json:"failed_party_ids"`
	FailedTagIDs      []string `json:"failed_tag_ids"`
}

// WeComAPIError represents an error from WeChat Work API
type WeComAPIError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (e *WeComAPIError) Error() string {
	return e.ErrMsg
}

// IsAccessTokenExpired checks if the error indicates expired access token
func (e *WeComAPIError) IsAccessTokenExpired() bool {
	return e.ErrCode == 40014 || e.ErrCode == 42001
}

// IsInvalidCredential checks if the error indicates invalid credential
func (e *WeComAPIError) IsInvalidCredential() bool {
	return e.ErrCode == 40013 || e.ErrCode == 40091
}
