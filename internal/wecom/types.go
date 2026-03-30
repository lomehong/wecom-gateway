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

// Meeting appointment types (Phase 1.3)

// CreateMeetingParams represents parameters for creating a meeting appointment
type CreateMeetingParams struct {
	Title         string            `json:"title"`
	StartDateTime time.Time         `json:"meeting_start_datetime"`
	Duration      int               `json:"meeting_duration"`          // seconds
	Invitees      *MeetingInvitees  `json:"invitees,omitempty"`
	MeetingType   int               `json:"meeting_type,omitempty"`    // 0=video, 1=voice, ...
	Settings      *MeetingSettings  `json:"settings,omitempty"`
}

// MeetingInvitees represents the invitees of a meeting
type MeetingInvitees struct {
	UserIDs []string `json:"userid,omitempty"`
	DeptIDs []string `json:"department,omitempty"`
}

// MeetingSettings represents meeting settings
type MeetingSettings struct {
	MuteUponEntry   bool `json:"mute_upon_entry,omitempty"`
	WaitingRoom     bool `json:"waiting_room,omitempty"`
	EnableRecording bool `json:"enable_recording,omitempty"`
}

// MeetingInfo represents information about a meeting
type MeetingInfo struct {
	MeetingID     string               `json:"meetingid"`
	Title         string               `json:"title"`
	Status        int                  `json:"status"`
	StartDateTime time.Time            `json:"meeting_start_datetime"`
	Duration      int                  `json:"meeting_duration"`
	Creator       string               `json:"creator"`
	Invitees      []MeetingInviteeInfo `json:"invitees,omitempty"`
	MeetingLink   string               `json:"meeting_link,omitempty"`
}

// MeetingInviteeInfo represents info about a meeting invitee
type MeetingInviteeInfo struct {
	UserID string `json:"userid"`
	Status int    `json:"status"`
}

// MeetingListOptions represents options for listing meetings
type MeetingListOptions struct {
	BeginDatetime string `json:"begin_datetime"`
	EndDatetime   string `json:"end_datetime"`
	Limit         int    `json:"limit,omitempty"`
	Cursor        string `json:"cursor,omitempty"`
}

// MeetingListResult represents the result of listing meetings
type MeetingListResult struct {
	Meetings   []MeetingInfo `json:"meetings"`
	NextCursor string        `json:"next_cursor"`
}

// Message pull types (Phase 3.1)

// ChatListResult represents the result of listing chats
type ChatListResult struct {
	ChatList []ChatInfo `json:"chat_list"`
}

// ChatInfo represents a chat conversation
type ChatInfo struct {
	ChatID    string `json:"chat_id"`
	Name      string `json:"name,omitempty"`
	ChatType  int    `json:"chat_type"`
	Owner     string `json:"owner,omitempty"`
	MemberCnt int    `json:"member_cnt,omitempty"`
}

// ChatMessagesResult represents the result of getting chat messages
type ChatMessagesResult struct {
	MsgList []ChatMessage `json:"msg_list"`
}

// ChatMessage represents a message in a chat
type ChatMessage struct {
	MsgID      string `json:"msgid"`
	MsgType    string `json:"msgtype"`
	From       string `json:"from"`
	ToList     string `json:"to_list"`
	ActionTime int64  `json:"action_time"`
	Text       struct {
		Content string `json:"content"`
	} `json:"text,omitempty"`
	Image struct {
		MediaID string `json:"media_id"`
	} `json:"image,omitempty"`
	File struct {
		MediaID string `json:"media_id"`
	} `json:"file,omitempty"`
	Voice struct {
		MediaID string `json:"media_id"`
	} `json:"voice,omitempty"`
	Video struct {
		MediaID string `json:"media_id"`
	} `json:"video,omitempty"`
	Link struct {
		Title    string `json:"title"`
		Desc     string `json:"desc"`
		LinkURL  string `json:"link_url"`
		ImageURL string `json:"image_url"`
	} `json:"link,omitempty"`
}

// Schedule availability types (Phase 3.2)

// AvailabilityOptions represents options for checking availability
type AvailabilityOptions struct {
	UserIDs   []string `json:"userids"`
	StartTime int64    `json:"start_time"`
	EndTime   int64    `json:"end_time"`
}

// UserAvailability represents the availability status of a user
type UserAvailability struct {
	UserID string          `json:"userid"`
	Slots  []AvailabilitySlot `json:"slots,omitempty"`
}

// AvailabilitySlot represents a time slot with availability info
type AvailabilitySlot struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
	BusyType  int   `json:"busy_type"`
}

// ContactUser represents a WeChat Work contact user
type ContactUser struct {
	UserID     string `json:"userid"`
	Name       string `json:"name"`
	Alias      string `json:"alias,omitempty"`
	Mobile     string `json:"mobile,omitempty"`
	Email      string `json:"email,omitempty"`
	Department []int  `json:"department,omitempty"`
	Position   string `json:"position,omitempty"`
	Gender     int    `json:"gender,omitempty"`
	Status     int    `json:"status,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
}

// ContactSimpleUser represents a simplified contact user from simplelist API
type ContactSimpleUser struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
	Dept   []int  `json:"department"`
}

// TodoListOptions represents options for querying todo list
type TodoListOptions struct {
	CreateBeginTime *time.Time `json:"create_begin_time,omitempty"`
	CreateEndTime   *time.Time `json:"create_end_time,omitempty"`
	RemindBeginTime *time.Time `json:"remind_begin_time,omitempty"`
	RemindEndTime   *time.Time `json:"remind_end_time,omitempty"`
	Limit           int        `json:"limit,omitempty"`
	Cursor          string     `json:"cursor,omitempty"`
}

// TodoListResult represents the result of listing todos
type TodoListResult struct {
	IndexList  []TodoIndex `json:"index_list"`
	NextCursor string      `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
}

// TodoIndex represents a todo index entry
type TodoIndex struct {
	TodoID     string     `json:"todo_id"`
	TodoStatus int        `json:"todo_status"`
	UserStatus int        `json:"user_status"`
	CreatorID  string     `json:"creator_id"`
	RemindTime *time.Time `json:"remind_time,omitempty"`
	CreateTime time.Time  `json:"create_time"`
	UpdateTime time.Time  `json:"update_time"`
}

// TodoDetail represents detailed todo information
type TodoDetail struct {
	TodoIndex
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Assignees []string `json:"assignees,omitempty"`
}

// CreateTodoParams represents parameters for creating a todo
type CreateTodoParams struct {
	Title      string     `json:"title,omitempty"`
	Content    string     `json:"content"`
	Assignees  []string   `json:"assignees,omitempty"`
	RemindTime *time.Time `json:"remind_time,omitempty"`
}

// UpdateTodoParams represents parameters for updating a todo
type UpdateTodoParams struct {
	Title      *string     `json:"title,omitempty"`
	Content    *string     `json:"content,omitempty"`
	Status     *int        `json:"status,omitempty"`      // 0=完成, 1=进行中
	RemindTime *time.Time  `json:"remind_time,omitempty"`
	Assignees  []string    `json:"assignees,omitempty"`
}
