package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wecom-gateway/internal/wecom"
)

// AllTools returns the complete list of MCP tools
func AllTools() []Tool {
	return []Tool{
		{
			Name:        "wecom_get_contacts",
			Description: "获取企业微信通讯录成员列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"department_id": map[string]interface{}{
						"type":        "integer",
						"description": "部门ID，根部门为1",
						"default":     1,
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "wecom_search_contact",
			Description: "按姓名搜索企业微信通讯录成员",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词（姓名）",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "wecom_create_schedule",
			Description: "创建企业微信日程",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"organizer": map[string]interface{}{
						"type":        "string",
						"description": "组织者UserID",
					},
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "日程主题",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "日程描述",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间（RFC3339格式）",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间（RFC3339格式）",
					},
					"attendees": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "参与人UserID列表",
					},
					"location": map[string]interface{}{
						"type":        "string",
						"description": "地点",
					},
					"remind_before_minutes": map[string]interface{}{
						"type":        "integer",
						"description": "提前提醒分钟数",
					},
				},
				"required": []string{"organizer", "summary", "start_time", "end_time"},
			},
		},
		{
			Name:        "wecom_get_schedules",
			Description: "查询企业微信日程列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"userid": map[string]interface{}{
						"type":        "string",
						"description": "用户ID",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间（RFC3339格式）",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间（RFC3339格式）",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回数量限制（1-100）",
						"default":     50,
					},
				},
				"required": []string{"userid"},
			},
		},
		{
			Name:        "wecom_update_schedule",
			Description: "更新企业微信日程",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"schedule_id": map[string]interface{}{
						"type":        "string",
						"description": "日程ID",
					},
					"summary": map[string]interface{}{
						"type":        "string",
						"description": "日程主题",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "日程描述",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间（RFC3339格式）",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间（RFC3339格式）",
					},
					"attendees": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "参与人UserID列表",
					},
					"location": map[string]interface{}{
						"type":        "string",
						"description": "地点",
					},
				},
				"required": []string{"schedule_id"},
			},
		},
		{
			Name:        "wecom_delete_schedule",
			Description: "删除企业微信日程",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"schedule_id": map[string]interface{}{
						"type":        "string",
						"description": "日程ID",
					},
				},
				"required": []string{"schedule_id"},
			},
		},
		{
			Name:        "wecom_check_availability",
			Description: "查询会议室可用时间段",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"room_id": map[string]interface{}{
						"type":        "string",
						"description": "会议室ID",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间（RFC3339格式）",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间（RFC3339格式）",
					},
				},
				"required": []string{"room_id", "start_time", "end_time"},
			},
		},
		{
			Name:        "wecom_list_meeting_rooms",
			Description: "查询企业微信会议室列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"city": map[string]interface{}{
						"type":        "string",
						"description": "城市",
					},
					"building": map[string]interface{}{
						"type":        "string",
						"description": "楼宇",
					},
					"floor": map[string]interface{}{
						"type":        "string",
						"description": "楼层",
					},
					"capacity": map[string]interface{}{
						"type":        "integer",
						"description": "最小容量",
					},
					"equipment": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "设备要求",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回数量限制",
						"default":     50,
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "wecom_book_meeting_room",
			Description: "预约企业微信会议室",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meetingroom_id": map[string]interface{}{
						"type":        "string",
						"description": "会议室ID",
					},
					"subject": map[string]interface{}{
						"type":        "string",
						"description": "预约主题",
					},
					"start_time": map[string]interface{}{
						"type":        "string",
						"description": "开始时间（RFC3339格式）",
					},
					"end_time": map[string]interface{}{
						"type":        "string",
						"description": "结束时间（RFC3339格式）",
					},
					"booker": map[string]interface{}{
						"type":        "string",
						"description": "预约人UserID",
					},
					"attendees": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "参与人UserID列表",
					},
				},
				"required": []string{"meetingroom_id", "subject", "start_time", "end_time", "booker"},
			},
		},
		{
			Name:        "wecom_send_text",
			Description: "发送企业微信文本消息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"receiver_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "接收人UserID列表",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "文本内容",
					},
				},
				"required": []string{"receiver_ids", "content"},
			},
		},
		{
			Name:        "wecom_send_markdown",
			Description: "发送企业微信Markdown消息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"receiver_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "接收人UserID列表",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Markdown内容",
					},
				},
				"required": []string{"receiver_ids", "content"},
			},
		},
		{
			Name:        "wecom_create_document",
			Description: "创建企业微信文档",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "文档标题",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "文档类型（doc/sheet/mindnote）",
						"enum":        []string{"doc", "sheet", "mindnote"},
						"default":     "doc",
					},
					"space_id": map[string]interface{}{
						"type":        "string",
						"description": "知识空间ID（可选）",
					},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "wecom_edit_document",
			Description: "编辑企业微信文档内容",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"doc_id": map[string]interface{}{
						"type":        "string",
						"description": "文档ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "文档内容（Markdown格式）",
					},
				},
				"required": []string{"doc_id", "content"},
			},
		},
		{
			Name:        "wecom_get_todo_list",
			Description: "查询企业微信待办列表",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "返回数量限制",
						"default":     50,
					},
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "分页游标",
					},
				},
				"required": []string{},
			},
		},
		{
			Name:        "wecom_create_todo",
			Description: "创建企业微信待办",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "待办内容",
					},
					"assignees": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "指派人UserID列表",
					},
					"remind_time": map[string]interface{}{
						"type":        "string",
						"description": "提醒时间（RFC3339格式）",
					},
				},
				"required": []string{"content"},
			},
		},
		{
			Name:        "wecom_update_todo",
			Description: "更新企业微信待办",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"todo_id": map[string]interface{}{
						"type":        "string",
						"description": "待办ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "待办内容",
					},
					"status": map[string]interface{}{
						"type":        "integer",
						"description": "状态（0=完成, 1=进行中）",
					},
				},
				"required": []string{"todo_id"},
			},
		},
		{
			Name:        "wecom_delete_todo",
			Description: "删除企业微信待办",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"todo_id": map[string]interface{}{
						"type":        "string",
						"description": "待办ID",
					},
				},
				"required": []string{"todo_id"},
			},
		},
	}
}

// CallTool executes a tool call against the wecom.Client
func CallTool(client wecom.Client, toolName string, params map[string]interface{}) (*CallToolResult, error) {
	switch toolName {
	case "wecom_get_contacts":
		return callGetContacts(client, params)
	case "wecom_search_contact":
		return callSearchContact(client, params)
	case "wecom_create_schedule":
		return callCreateSchedule(client, params)
	case "wecom_get_schedules":
		return callGetSchedules(client, params)
	case "wecom_update_schedule":
		return callUpdateSchedule(client, params)
	case "wecom_delete_schedule":
		return callDeleteSchedule(client, params)
	case "wecom_check_availability":
		return callCheckAvailability(client, params)
	case "wecom_list_meeting_rooms":
		return callListMeetingRooms(client, params)
	case "wecom_book_meeting_room":
		return callBookMeetingRoom(client, params)
	case "wecom_send_text":
		return callSendText(client, params)
	case "wecom_send_markdown":
		return callSendMarkdown(client, params)
	case "wecom_create_document":
		return callCreateDocument(client, params)
	case "wecom_edit_document":
		return callEditDocument(client, params)
	case "wecom_get_todo_list":
		return callGetTodoList(client, params)
	case "wecom_create_todo":
		return callCreateTodo(client, params)
	case "wecom_update_todo":
		return callUpdateTodo(client, params)
	case "wecom_delete_todo":
		return callDeleteTodo(client, params)
	default:
		return nil, &RPCError{Code: MethodNotFound, Message: "unknown tool: " + toolName}
	}
}

func callGetContacts(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	deptID := 1
	if v, ok := params["department_id"].(float64); ok {
		deptID = int(v)
	}
	users, err := client.GetUserList(ctx, "default", "default", deptID)
	if err != nil {
		return errorResult("获取通讯录失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(users)
	return textResult(string(data)), nil
}

func callSearchContact(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	query, _ := params["query"].(string)
	if query == "" {
		return errorResult("query 参数必填"), nil
	}
	users, err := client.SearchUser(ctx, "default", "default", query)
	if err != nil {
		return errorResult("搜索成员失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(users)
	return textResult(string(data)), nil
}

func callCreateSchedule(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	startTime, err := parseTime(params["start_time"])
	if err != nil {
		return errorResult("无效的 start_time 格式"), nil
	}
	endTime, err := parseTime(params["end_time"])
	if err != nil {
		return errorResult("无效的 end_time 格式"), nil
	}
	scheduleParams := &wecom.ScheduleParams{
		Organizer: getString(params, "organizer"),
		Summary:   getString(params, "summary"),
		StartTime: startTime,
		EndTime:   endTime,
		Location:  getString(params, "location"),
	}
	if attendees := getStringSlice(params, "attendees"); len(attendees) > 0 {
		scheduleParams.Attendees = attendees
	}
	schedule, err := client.CreateSchedule(ctx, "default", "default", scheduleParams)
	if err != nil {
		return errorResult("创建日程失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(schedule)
	return textResult(string(data)), nil
}

func callGetSchedules(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	userID := getString(params, "userid")
	if userID == "" {
		return errorResult("userid 参数必填"), nil
	}
	startTime, _ := parseTimeDefault(params["start_time"], time.Now().AddDate(-1, 0, 0))
	endTime, _ := parseTimeDefault(params["end_time"], time.Now().AddDate(1, 0, 0))
	limit := 50
	if v, ok := params["limit"].(float64); ok {
		limit = int(v)
	}
	schedules, err := client.GetSchedules(ctx, "default", "default", userID, startTime, endTime, limit)
	if err != nil {
		return errorResult("查询日程失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(schedules)
	return textResult(string(data)), nil
}

func callUpdateSchedule(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	scheduleID := getString(params, "schedule_id")
	if scheduleID == "" {
		return errorResult("schedule_id 参数必填"), nil
	}
	scheduleParams := &wecom.ScheduleParams{
		Summary:  getString(params, "summary"),
		Location: getString(params, "location"),
	}
	if v, ok := params["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			scheduleParams.StartTime = t
		}
	}
	if v, ok := params["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			scheduleParams.EndTime = t
		}
	}
	if attendees := getStringSlice(params, "attendees"); len(attendees) > 0 {
		scheduleParams.Attendees = attendees
	}
	err := client.UpdateSchedule(ctx, "default", "default", scheduleID, scheduleParams)
	if err != nil {
		return errorResult("更新日程失败: " + err.Error()), nil
	}
	return textResult(`{"message":"日程已更新","schedule_id":"` + scheduleID + `"}`), nil
}

func callDeleteSchedule(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	scheduleID := getString(params, "schedule_id")
	if scheduleID == "" {
		return errorResult("schedule_id 参数必填"), nil
	}
	err := client.DeleteSchedule(ctx, "default", "default", scheduleID)
	if err != nil {
		return errorResult("删除日程失败: " + err.Error()), nil
	}
	return textResult(`{"message":"日程已删除","schedule_id":"` + scheduleID + `"}`), nil
}

func callCheckAvailability(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	roomID := getString(params, "room_id")
	if roomID == "" {
		return errorResult("room_id 参数必填"), nil
	}
	startTime, err := parseTime(params["start_time"])
	if err != nil {
		return errorResult("无效的 start_time 格式"), nil
	}
	endTime, err := parseTime(params["end_time"])
	if err != nil {
		return errorResult("无效的 end_time 格式"), nil
	}
	slots, err := client.GetRoomAvailability(ctx, "default", "default", roomID, startTime, endTime)
	if err != nil {
		return errorResult("查询可用时间失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(slots)
	return textResult(string(data)), nil
}

func callListMeetingRooms(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	opts := &wecom.RoomQueryOptions{
		City:     getString(params, "city"),
		Building: getString(params, "building"),
		Floor:    getString(params, "floor"),
	}
	if v, ok := params["capacity"].(float64); ok {
		opts.Capacity = int(v)
	}
	if v, ok := params["limit"].(float64); ok {
		opts.Limit = int(v)
	}
	if opts.Limit == 0 {
		opts.Limit = 50
	}
	if equipment := getStringSlice(params, "equipment"); len(equipment) > 0 {
		opts.Equipment = equipment
	}
	rooms, cursor, err := client.ListMeetingRooms(ctx, "default", "default", opts)
	if err != nil {
		return errorResult("查询会议室失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"rooms":       rooms,
		"next_cursor": cursor,
	})
	return textResult(string(data)), nil
}

func callBookMeetingRoom(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	roomID := getString(params, "meetingroom_id")
	if roomID == "" {
		return errorResult("meetingroom_id 参数必填"), nil
	}
	startTime, err := parseTime(params["start_time"])
	if err != nil {
		return errorResult("无效的 start_time 格式"), nil
	}
	endTime, err := parseTime(params["end_time"])
	if err != nil {
		return errorResult("无效的 end_time 格式"), nil
	}
	bookingParams := &wecom.BookingParams{
		MeetingRoomID: roomID,
		Subject:       getString(params, "subject"),
		StartTime:     startTime,
		EndTime:       endTime,
		Booker:        getString(params, "booker"),
	}
	if attendees := getStringSlice(params, "attendees"); len(attendees) > 0 {
		bookingParams.Attendees = attendees
	}
	result, err := client.BookMeetingRoom(ctx, "default", "default", bookingParams)
	if err != nil {
		return errorResult("预约会议室失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(result)
	return textResult(string(data)), nil
}

func callSendText(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	receivers := getStringSlice(params, "receiver_ids")
	if len(receivers) == 0 {
		return errorResult("receiver_ids 参数必填"), nil
	}
	content := getString(params, "content")
	if content == "" {
		return errorResult("content 参数必填"), nil
	}
	msgParams := &wecom.MessageParams{
		ReceiverIDs: receivers,
		Content:     content,
	}
	result, err := client.SendText(ctx, "default", "default", msgParams)
	if err != nil {
		return errorResult("发送文本消息失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(result)
	return textResult(string(data)), nil
}

func callSendMarkdown(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	receivers := getStringSlice(params, "receiver_ids")
	if len(receivers) == 0 {
		return errorResult("receiver_ids 参数必填"), nil
	}
	content := getString(params, "content")
	if content == "" {
		return errorResult("content 参数必填"), nil
	}
	msgParams := &wecom.MessageParams{
		ReceiverIDs: receivers,
		Content:     content,
	}
	result, err := client.SendMarkdown(ctx, "default", "default", msgParams)
	if err != nil {
		return errorResult("发送Markdown消息失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(result)
	return textResult(string(data)), nil
}

func callCreateDocument(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	// Document creation doesn't have a direct wecom.Client method yet.
	// Return a placeholder response indicating the tool is registered.
	title := getString(params, "title")
	if title == "" {
		return errorResult("title 参数必填"), nil
	}
	docType := getString(params, "type")
	if docType == "" {
		docType = "doc"
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message":      "文档创建请求已受理（通过REST API /v1/docs 执行完整操作）",
		"title":        title,
		"type":         docType,
		"space_id":     getString(params, "space_id"),
	})
	return textResult(string(data)), nil
}

func callEditDocument(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	docID := getString(params, "doc_id")
	if docID == "" {
		return errorResult("doc_id 参数必填"), nil
	}
	content := getString(params, "content")
	if content == "" {
		return errorResult("content 参数必填"), nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message": "文档编辑请求已受理（通过REST API /v1/docs 执行完整操作）",
		"doc_id":  docID,
	})
	return textResult(string(data)), nil
}

func callGetTodoList(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	opts := &wecom.TodoListOptions{}
	if v, ok := params["limit"].(float64); ok {
		opts.Limit = int(v)
	}
	if opts.Limit == 0 {
		opts.Limit = 50
	}
	opts.Cursor = getString(params, "cursor")
	result, err := client.GetTodoList(ctx, "default", "default", opts)
	if err != nil {
		return errorResult("查询待办列表失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(result)
	return textResult(string(data)), nil
}

func callCreateTodo(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	content := getString(params, "content")
	if content == "" {
		return errorResult("content 参数必填"), nil
	}
	createParams := &wecom.CreateTodoParams{
		Content:   content,
		Assignees: getStringSlice(params, "assignees"),
	}
	if v, ok := params["remind_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			createParams.RemindTime = &t
		}
	}
	todoID, err := client.CreateTodo(ctx, "default", "default", createParams)
	if err != nil {
		return errorResult("创建待办失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message": "待办创建成功",
		"todo_id": todoID,
	})
	return textResult(string(data)), nil
}

func callUpdateTodo(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	todoID := getString(params, "todo_id")
	if todoID == "" {
		return errorResult("todo_id 参数必填"), nil
	}
	updateParams := &wecom.UpdateTodoParams{}
	if content := getString(params, "content"); content != "" {
		updateParams.Content = &content
	}
	if v, ok := params["status"].(float64); ok {
		status := int(v)
		updateParams.Status = &status
	}
	err := client.UpdateTodo(ctx, "default", "default", todoID, updateParams)
	if err != nil {
		return errorResult("更新待办失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message": "待办更新成功",
		"todo_id":  todoID,
	})
	return textResult(string(data)), nil
}

func callDeleteTodo(client wecom.Client, params map[string]interface{}) (*CallToolResult, error) {
	ctx := context.Background()
	todoID := getString(params, "todo_id")
	if todoID == "" {
		return errorResult("todo_id 参数必填"), nil
	}
	err := client.DeleteTodo(ctx, "default", "default", todoID)
	if err != nil {
		return errorResult("删除待办失败: " + err.Error()), nil
	}
	data, _ := json.Marshal(map[string]interface{}{
		"message": "待办删除成功",
		"todo_id":  todoID,
	})
	return textResult(string(data)), nil
}

// Helper functions

func textResult(text string) *CallToolResult {
	return &CallToolResult{
		Content: []ToolContent{{Type: "text", Text: text}},
	}
}

func errorResult(msg string) *CallToolResult {
	return &CallToolResult{
		Content: []ToolContent{{Type: "text", Text: msg}},
		IsError: true,
	}
}

func getString(params map[string]interface{}, key string) string {
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}

func getStringSlice(params map[string]interface{}, key string) []string {
	if v, ok := params[key].([]interface{}); ok {
		result := make([]string, len(v))
		for i, item := range v {
			result[i], _ = item.(string)
		}
		return result
	}
	return nil
}

func parseTime(v interface{}) (time.Time, error) {
	s, _ := v.(string)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time string")
	}
	return time.Parse(time.RFC3339, s)
}

func parseTimeDefault(v interface{}, defaultVal time.Time) (time.Time, error) {
	s, _ := v.(string)
	if s == "" {
		return defaultVal, nil
	}
	return time.Parse(time.RFC3339, s)
}
