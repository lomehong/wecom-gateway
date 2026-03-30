package openapi

import "encoding/json"

// Spec represents the complete OpenAPI 3.0 specification
type Spec struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components"`
	Tags       []Tag                 `json:"tags"`
}

// Info contains API metadata
type Info struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	Version        string `json:"version"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// License information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents an API server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Tag represents an API tag
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PathItem represents a path item
type PathItem map[string]Operation

// Operation represents an HTTP operation
type Operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	OperationID string              `json:"operationId,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter represents an operation parameter
type Parameter struct {
	Name            string          `json:"name"`
	In              string          `json:"in"`
	Description     string          `json:"description,omitempty"`
	Required        bool            `json:"required,omitempty"`
	Schema          *SchemaRef      `json:"schema,omitempty"`
	Style           string          `json:"style,omitempty"`
	Explode         bool            `json:"explode,omitempty"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string              `json:"description,omitempty"`
	Required    bool                `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType represents a media type
type MediaType struct {
	Schema       *SchemaRef   `json:"schema"`
	Example      interface{}  `json:"example,omitempty"`
	Examples     map[string]Example `json:"examples,omitempty"`
}

// Example represents an example
type Example struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Value       interface{} `json:"value"`
}

// Response represents an HTTP response
type Response struct {
	Description string              `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// Components holds reusable schemas
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
	Schemas         map[string]SchemaRef      `json:"schemas"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	In          string `json:"in,omitempty"`
	Name        string `json:"name,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// SchemaRef represents a schema or schema reference
type SchemaRef struct {
	Ref         string             `json:"$ref,omitempty"`
	Type        string             `json:"type,omitempty"`
	Format      string             `json:"format,omitempty"`
	Description string             `json:"description,omitempty"`
	Properties  map[string]SchemaRef `json:"properties,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Items       *SchemaRef         `json:"items,omitempty"`
	Enum        []interface{}      `json:"enum,omitempty"`
	Default     interface{}        `json:"default,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
	AdditionalProperties *SchemaRef `json:"additionalProperties,omitempty"`
}

// GetSpec returns the complete OpenAPI 3.0 specification
func GetSpec() *Spec {
	return &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       "WeCom Gateway API",
			Description: "企业微信 AI Agent 网关 — 统一的企业微信能力接口，支持日程管理、会议室预约、消息发送、文档管理、通讯录查询、待办管理等功能。",
			Version:     "1.0.0",
			Contact: &Contact{
				Name:  "WeCom Gateway Team",
				Email: "team@wecom-gateway.dev",
			},
			License: &License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
		Servers: []Server{
			{
				URL:         "http://localhost:8080",
				Description: "本地开发服务器",
			},
		},
		Tags: []Tag{
			{Name: "schedules", Description: "日程管理"},
			{Name: "meeting-rooms", Description: "会议室管理"},
			{Name: "meetings", Description: "会议预约"},
			{Name: "messages", Description: "消息发送"},
			{Name: "documents", Description: "文档管理"},
			{Name: "sheets", Description: "智能表格"},
			{Name: "contacts", Description: "通讯录查询"},
			{Name: "todos", Description: "待办管理"},
			{Name: "admin", Description: "管理接口"},
			{Name: "mcp", Description: "MCP 协议"},
		},
		Paths: buildPaths(),
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "API Key",
					Description:  "使用 API Key (wgk_*) 作为 Bearer Token 进行认证",
				},
			},
			Schemas: buildSchemas(),
		},
	}
}

// GetSpecJSON returns the OpenAPI spec as JSON bytes
func GetSpecJSON() ([]byte, error) {
	spec := GetSpec()
	return json.MarshalIndent(spec, "", "  ")
}

func buildPaths() map[string]PathItem {
	paths := make(map[string]PathItem)

	// === Schedules ===
	paths["/v1/schedules"] = PathItem{
		"post": Operation{
			Tags:        []string{"schedules"},
			Summary:     "创建日程",
			OperationID: "createSchedule",
			RequestBody: jsonBody("日程参数", true, "#/components/schemas/CreateScheduleRequest"),
			Responses: map[string]Response{
				"201": response("创建成功", "#/components/schemas/Schedule"),
				"400": errResponse("参数错误"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"get": Operation{
			Tags:        []string{"schedules"},
			Summary:     "查询日程列表",
			OperationID: "getSchedules",
			Parameters: []Parameter{
				{Name: "userid", In: "query", Description: "用户ID", Required: true, Schema: strSchema()},
				{Name: "start_time", In: "query", Description: "开始时间 (RFC3339)", Schema: strSchema()},
				{Name: "end_time", In: "query", Description: "结束时间 (RFC3339)", Schema: strSchema()},
				{Name: "limit", In: "query", Description: "数量限制 (1-100)", Schema: intSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ScheduleList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/schedules/{id}"] = PathItem{
		"get": Operation{
			Tags:        []string{"schedules"},
			Summary:     "获取日程详情",
			OperationID: "getScheduleByID",
			Parameters: []Parameter{pathParam("id", "日程ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/Schedule"),
				"404": errResponse("日程不存在"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"patch": Operation{
			Tags:        []string{"schedules"},
			Summary:     "更新日程",
			OperationID: "updateSchedule",
			Parameters: []Parameter{pathParam("id", "日程ID")},
			RequestBody: jsonBody("更新参数", false, "#/components/schemas/UpdateScheduleRequest"),
			Responses: map[string]Response{
				"200": response("更新成功", "#/components/schemas/ApiResponse"),
				"404": errResponse("日程不存在"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"delete": Operation{
			Tags:        []string{"schedules"},
			Summary:     "删除日程",
			OperationID: "deleteSchedule",
			Parameters: []Parameter{pathParam("id", "日程ID")},
			Responses: map[string]Response{
				"200": response("删除成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Meeting Rooms ===
	paths["/v1/meeting-rooms"] = PathItem{
		"get": Operation{
			Tags:        []string{"meeting-rooms"},
			Summary:     "查询会议室列表",
			OperationID: "listMeetingRooms",
			Parameters: []Parameter{
				{Name: "city", In: "query", Description: "城市", Schema: strSchema()},
				{Name: "building", In: "query", Description: "楼宇", Schema: strSchema()},
				{Name: "floor", In: "query", Description: "楼层", Schema: strSchema()},
				{Name: "capacity", In: "query", Description: "最小容量", Schema: intSchema()},
				{Name: "equipment", In: "query", Description: "设备要求", Schema: arrSchema(strSchema())},
				{Name: "limit", In: "query", Description: "数量限制", Schema: intSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/MeetingRoomList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/meeting-rooms/{id}/availability"] = PathItem{
		"get": Operation{
			Tags:        []string{"meeting-rooms"},
			Summary:     "查询会议室可用时间",
			OperationID: "getRoomAvailability",
			Parameters: []Parameter{
				pathParam("id", "会议室ID"),
				{Name: "start_time", In: "query", Description: "开始时间 (RFC3339)", Required: true, Schema: strSchema()},
				{Name: "end_time", In: "query", Description: "结束时间 (RFC3339)", Required: true, Schema: strSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/TimeSlotList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/meeting-rooms/{id}/bookings"] = PathItem{
		"post": Operation{
			Tags:        []string{"meeting-rooms"},
			Summary:     "预约会议室",
			OperationID: "bookMeetingRoom",
			Parameters: []Parameter{pathParam("id", "会议室ID")},
			RequestBody: jsonBody("预约参数", true, "#/components/schemas/BookMeetingRoomRequest"),
			Responses: map[string]Response{
				"201": response("预约成功", "#/components/schemas/BookingResult"),
				"409": errResponse("时间段冲突"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Meetings (Phase 1.3) ===
	paths["/v1/meetings"] = PathItem{
		"post": Operation{
			Tags:        []string{"meetings"},
			Summary:     "创建预约会议",
			OperationID: "createMeeting",
			RequestBody: jsonBody("会议参数", true, "#/components/schemas/CreateMeetingRequest"),
			Responses: map[string]Response{
				"201": response("创建成功", "#/components/schemas/MeetingInfo"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"get": Operation{
			Tags:        []string{"meetings"},
			Summary:     "查询会议列表",
			OperationID: "listMeetings",
			Parameters: []Parameter{
				{Name: "begin_datetime", In: "query", Description: "开始时间", Schema: strSchema()},
				{Name: "end_datetime", In: "query", Description: "结束时间", Schema: strSchema()},
				{Name: "limit", In: "query", Description: "数量限制", Schema: intSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/MeetingList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/meetings/{id}"] = PathItem{
		"get": Operation{
			Tags:        []string{"meetings"},
			Summary:     "获取会议详情",
			OperationID: "getMeetingInfo",
			Parameters: []Parameter{pathParam("id", "会议ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/MeetingInfo"),
				"404": errResponse("会议不存在"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"delete": Operation{
			Tags:        []string{"meetings"},
			Summary:     "取消会议",
			OperationID: "cancelMeeting",
			Parameters: []Parameter{pathParam("id", "会议ID")},
			Responses: map[string]Response{
				"200": response("取消成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Messages ===
	paths["/v1/messages/text"] = PathItem{
		"post": Operation{
			Tags:        []string{"messages"},
			Summary:     "发送文本消息",
			OperationID: "sendText",
			RequestBody: jsonBody("消息参数", true, "#/components/schemas/SendTextRequest"),
			Responses: map[string]Response{
				"200": response("发送成功", "#/components/schemas/SendResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/messages/markdown"] = PathItem{
		"post": Operation{
			Tags:        []string{"messages"},
			Summary:     "发送 Markdown 消息",
			OperationID: "sendMarkdown",
			RequestBody: jsonBody("消息参数", true, "#/components/schemas/SendTextRequest"),
			Responses: map[string]Response{
				"200": response("发送成功", "#/components/schemas/SendResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/messages/image"] = PathItem{
		"post": Operation{
			Tags:        []string{"messages"},
			Summary:     "发送图片消息",
			OperationID: "sendImage",
			RequestBody: jsonBody("图片参数", true, "#/components/schemas/SendImageRequest"),
			Responses: map[string]Response{
				"200": response("发送成功", "#/components/schemas/SendResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/messages/file"] = PathItem{
		"post": Operation{
			Tags:        []string{"messages"},
			Summary:     "发送文件消息",
			OperationID: "sendFile",
			RequestBody: jsonBody("文件参数", true, "#/components/schemas/SendFileRequest"),
			Responses: map[string]Response{
				"200": response("发送成功", "#/components/schemas/SendResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/messages/card"] = PathItem{
		"post": Operation{
			Tags:        []string{"messages"},
			Summary:     "发送卡片消息",
			OperationID: "sendCard",
			RequestBody: jsonBody("卡片参数", true, "#/components/schemas/SendCardRequest"),
			Responses: map[string]Response{
				"200": response("发送成功", "#/components/schemas/SendResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Documents ===
	paths["/v1/docs"] = PathItem{
		"get": Operation{
			Tags:        []string{"documents"},
			Summary:     "查询文档列表",
			OperationID: "listDocuments",
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/DocumentList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"post": Operation{
			Tags:        []string{"documents"},
			Summary:     "创建文档",
			OperationID: "createDocument",
			RequestBody: jsonBody("文档参数", true, "#/components/schemas/CreateDocumentRequest"),
			Responses: map[string]Response{
				"201": response("创建成功", "#/components/schemas/DocumentInfo"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/docs/{docid}"] = PathItem{
		"get": Operation{
			Tags:        []string{"documents"},
			Summary:     "获取文档详情",
			OperationID: "getDocument",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/DocumentInfo"),
				"404": errResponse("文档不存在"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"delete": Operation{
			Tags:        []string{"documents"},
			Summary:     "删除文档",
			OperationID: "deleteDocument",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			Responses: map[string]Response{
				"200": response("删除成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/docs/{docid}/rename"] = PathItem{
		"put": Operation{
			Tags:        []string{"documents"},
			Summary:     "重命名文档",
			OperationID: "renameDocument",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			RequestBody: jsonBody("重命名参数", true, "#/components/schemas/RenameDocumentRequest"),
			Responses: map[string]Response{
				"200": response("重命名成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/docs/{docid}/content"] = PathItem{
		"put": Operation{
			Tags:        []string{"documents"},
			Summary:     "编辑文档内容",
			OperationID: "editDocumentContent",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			RequestBody: jsonBody("内容参数", true, "#/components/schemas/EditContentRequest"),
			Responses: map[string]Response{
				"200": response("编辑成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/docs/{docid}/share"] = PathItem{
		"post": Operation{
			Tags:        []string{"documents"},
			Summary:     "分享文档",
			OperationID: "shareDocument",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			RequestBody: jsonBody("分享参数", true, "#/components/schemas/ShareDocumentRequest"),
			Responses: map[string]Response{
				"200": response("分享成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/docs/{docid}/permissions"] = PathItem{
		"get": Operation{
			Tags:        []string{"documents"},
			Summary:     "获取文档权限",
			OperationID: "getDocumentPermissions",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Sheets (Phase 3.3) ===
	paths["/v1/sheets"] = PathItem{
		"post": Operation{
			Tags:        []string{"sheets"},
			Summary:     "创建智能表格",
			OperationID: "createSheet",
			RequestBody: jsonBody("表格参数", true, "#/components/schemas/CreateSheetRequest"),
			Responses: map[string]Response{
				"201": response("创建成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/sheets/{docid}/sheets"] = PathItem{
		"get": Operation{
			Tags:        []string{"sheets"},
			Summary:     "查询子表列表",
			OperationID: "listSheetTabs",
			Parameters: []Parameter{pathParam("docid", "文档ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/sheets/{docid}/sheets/{sheetid}/records"] = PathItem{
		"get": Operation{
			Tags:        []string{"sheets"},
			Summary:     "查询记录",
			OperationID: "listSheetRecords",
			Parameters: []Parameter{
				pathParam("docid", "文档ID"),
				pathParam("sheetid", "子表ID"),
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"post": Operation{
			Tags:        []string{"sheets"},
			Summary:     "添加记录",
			OperationID: "createSheetRecord",
			Parameters: []Parameter{
				pathParam("docid", "文档ID"),
				pathParam("sheetid", "子表ID"),
			},
			RequestBody: jsonBody("记录数据", true, "#/components/schemas/ApiResponse"),
			Responses: map[string]Response{
				"201": response("添加成功", "#/components/schemas/ApiResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Contacts (Phase 1.1) ===
	paths["/v1/contacts/users"] = PathItem{
		"get": Operation{
			Tags:        []string{"contacts"},
			Summary:     "获取可见范围成员列表",
			OperationID: "getContactUserList",
			Parameters: []Parameter{
				{Name: "department_id", In: "query", Description: "部门ID，根部门为1", Schema: intSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ContactUserList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/contacts/users/search"] = PathItem{
		"get": Operation{
			Tags:        []string{"contacts"},
			Summary:     "按姓名搜索成员",
			OperationID: "searchContact",
			Parameters: []Parameter{
				{Name: "query", In: "query", Description: "搜索关键词", Required: true, Schema: strSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/ContactUserList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Todos (Phase 1.2) ===
	paths["/v1/todos"] = PathItem{
		"get": Operation{
			Tags:        []string{"todos"},
			Summary:     "查询待办列表",
			OperationID: "getTodoList",
			Parameters: []Parameter{
				{Name: "limit", In: "query", Description: "数量限制", Schema: intSchema()},
				{Name: "cursor", In: "query", Description: "分页游标", Schema: strSchema()},
			},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/TodoList"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"post": Operation{
			Tags:        []string{"todos"},
			Summary:     "创建待办",
			OperationID: "createTodo",
			RequestBody: jsonBody("待办参数", true, "#/components/schemas/CreateTodoRequest"),
			Responses: map[string]Response{
				"201": response("创建成功", "#/components/schemas/TodoCreateResult"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/todos/{id}"] = PathItem{
		"get": Operation{
			Tags:        []string{"todos"},
			Summary:     "查询待办详情",
			OperationID: "getTodoDetail",
			Parameters: []Parameter{pathParam("id", "待办ID")},
			Responses: map[string]Response{
				"200": response("查询成功", "#/components/schemas/TodoDetail"),
				"404": errResponse("待办不存在"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"put": Operation{
			Tags:        []string{"todos"},
			Summary:     "更新待办",
			OperationID: "updateTodo",
			Parameters: []Parameter{pathParam("id", "待办ID")},
			RequestBody: jsonBody("更新参数", false, "#/components/schemas/UpdateTodoRequest"),
			Responses: map[string]Response{
				"200": response("更新成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
		"delete": Operation{
			Tags:        []string{"todos"},
			Summary:     "删除待办",
			OperationID: "deleteTodo",
			Parameters: []Parameter{pathParam("id", "待办ID")},
			Responses: map[string]Response{
				"200": response("删除成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}
	paths["/v1/todos/{id}/status"] = PathItem{
		"put": Operation{
			Tags:        []string{"todos"},
			Summary:     "变更待办状态",
			OperationID: "changeTodoStatus",
			Parameters: []Parameter{pathParam("id", "待办ID")},
			RequestBody: jsonBody("状态参数", true, "#/components/schemas/ChangeTodoStatusRequest"),
			Responses: map[string]Response{
				"200": response("状态更新成功", "#/components/schemas/ApiResponse"),
				"403": errResponse("权限不足"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === Admin ===
	paths["/v1/admin/login"] = PathItem{
		"post": Operation{
			Tags:        []string{"admin"},
			Summary:     "管理员登录",
			OperationID: "adminLogin",
			RequestBody: jsonBody("登录参数", true, "#/components/schemas/LoginRequest"),
			Responses: map[string]Response{
				"200": response("登录成功", "#/components/schemas/LoginResponse"),
				"401": errResponse("认证失败"),
			},
		},
	}
	paths["/v1/admin/api-keys"] = PathItem{
		"get": Operation{
			Tags:        []string{"admin"},
			Summary:     "获取 API Key 列表",
			OperationID: "listAPIKeys",
			Responses:   map[string]Response{"200": response("查询成功", "#/components/schemas/ApiResponse")},
			Security:    []map[string][]string{{"BearerAuth": {}}},
		},
		"post": Operation{
			Tags:        []string{"admin"},
			Summary:     "创建 API Key",
			OperationID: "createAPIKey",
			RequestBody: jsonBody("API Key 参数", true, "#/components/schemas/ApiResponse"),
			Responses:   map[string]Response{"201": response("创建成功", "#/components/schemas/ApiResponse")},
			Security:    []map[string][]string{{"BearerAuth": {}}},
		},
	}

	// === MCP ===
	paths["/mcp"] = PathItem{
		"get": Operation{
			Tags:        []string{"mcp"},
			Summary:     "MCP SSE 端点",
			Description: "Server-Sent Events 连接，用于 MCP 流式响应",
			OperationID: "mcpSSE",
			Responses: map[string]Response{
				"200": Response{Description: "SSE 流"},
			},
		},
		"post": Operation{
			Tags:        []string{"mcp"},
			Summary:     "MCP JSON-RPC 2.0 端点",
			Description: "处理 MCP 协议的 JSON-RPC 请求",
			OperationID: "mcpRPC",
			RequestBody: jsonBody("JSON-RPC 请求", true, "#/components/schemas/MCPRequest"),
			Responses: map[string]Response{
				"200": response("JSON-RPC 响应", "#/components/schemas/MCPResponse"),
			},
			Security: []map[string][]string{{"BearerAuth": {}}},
		},
	}

	return paths
}

func buildSchemas() map[string]SchemaRef {
	return map[string]SchemaRef{
		// Common
		"ApiResponse": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Description: "业务状态码", Example: 0},
				"message": {Type: "string", Description: "状态消息", Example: "ok"},
				"data":    {Type: "object", Description: "响应数据"},
			},
		},
		"ErrorResponse": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Description: "错误码"},
				"message": {Type: "string", Description: "错误消息"},
			},
		},

		// Schedule
		"CreateScheduleRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"organizer":              {Type: "string", Description: "组织者UserID"},
				"summary":                {Type: "string", Description: "日程主题"},
				"description":            {Type: "string", Description: "日程描述"},
				"start_time":             {Type: "string", Format: "date-time", Description: "开始时间"},
				"end_time":               {Type: "string", Format: "date-time", Description: "结束时间"},
				"attendees":              {Type: "array", Items: strSchema(), Description: "参与人UserID列表"},
				"location":               {Type: "string", Description: "地点"},
				"remind_before_minutes":  {Type: "integer", Description: "提前提醒分钟数"},
			},
			Required: []string{"organizer", "summary", "start_time", "end_time"},
		},
		"UpdateScheduleRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"summary":                {Type: "string", Description: "日程主题"},
				"description":            {Type: "string", Description: "日程描述"},
				"start_time":             {Type: "string", Format: "date-time", Description: "开始时间"},
				"end_time":               {Type: "string", Format: "date-time", Description: "结束时间"},
				"attendees":              {Type: "array", Items: strSchema(), Description: "参与人UserID列表"},
				"location":               {Type: "string", Description: "地点"},
				"remind_before_minutes":  {Type: "integer", Description: "提前提醒分钟数"},
			},
		},
		"Schedule": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"schedule_id":   {Type: "string", Description: "日程ID"},
				"organizer":     {Type: "string", Description: "组织者"},
				"summary":       {Type: "string", Description: "主题"},
				"description":   {Type: "string", Description: "描述"},
				"start_time":    {Type: "string", Format: "date-time", Description: "开始时间"},
				"end_time":      {Type: "string", Format: "date-time", Description: "结束时间"},
				"attendees":     {Type: "array", Items: strSchema(), Description: "参与人"},
				"location":      {Type: "string", Description: "地点"},
				"cal_id":        {Type: "string", Description: "日历ID"},
			},
		},
		"ScheduleList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":      {Type: "integer", Example: 0},
				"message":   {Type: "string", Example: "ok"},
				"data": {Type: "object", Properties: map[string]SchemaRef{
					"schedules": {Type: "array", Items: schemaRef("#/components/schemas/Schedule")},
					"count":     {Type: "integer"},
				}},
			},
		},

		// Meeting Room
		"BookMeetingRoomRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"meetingroom_id": {Type: "string", Description: "会议室ID"},
				"subject":        {Type: "string", Description: "预约主题"},
				"start_time":     {Type: "string", Format: "date-time", Description: "开始时间"},
				"end_time":       {Type: "string", Format: "date-time", Description: "结束时间"},
				"booker":         {Type: "string", Description: "预约人UserID"},
				"attendees":      {Type: "array", Items: strSchema(), Description: "参与人"},
			},
			Required: []string{"meetingroom_id", "subject", "start_time", "end_time", "booker"},
		},
		"MeetingRoom": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"mbooking_id": {Type: "string", Description: "会议室ID"},
				"name":        {Type: "string", Description: "会议室名称"},
				"capacity":    {Type: "integer", Description: "容量"},
				"city":        {Type: "string", Description: "城市"},
				"building":    {Type: "string", Description: "楼宇"},
				"floor":       {Type: "string", Description: "楼层"},
				"equipment":   {Type: "array", Items: strSchema(), Description: "设备"},
			},
		},
		"MeetingRoomList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code": {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data": {Type: "object", Properties: map[string]SchemaRef{
					"rooms":       {Type: "array", Items: schemaRef("#/components/schemas/MeetingRoom")},
					"count":       {Type: "integer"},
					"next_cursor": {Type: "string"},
				}},
			},
		},
		"TimeSlotList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code": {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data": {Type: "object", Properties: map[string]SchemaRef{
					"room_id": {Type: "string"},
					"slots": {Type: "array", Items: schemaRef("#/components/schemas/TimeSlot")},
					"count": {Type: "integer"},
				}},
			},
		},
		"TimeSlot": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"start_time": {Type: "string", Format: "date-time"},
				"end_time":   {Type: "string", Format: "date-time"},
			},
		},
		"BookingResult": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"booking_id":  {Type: "string"},
				"schedule_id": {Type: "string"},
				"start_time":  {Type: "string", Format: "date-time"},
				"end_time":    {Type: "string", Format: "date-time"},
			},
		},

		// Meeting
		"CreateMeetingRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"title":                  {Type: "string", Description: "会议主题"},
				"meeting_start_datetime":  {Type: "string", Format: "date-time", Description: "开始时间"},
				"meeting_duration":        {Type: "integer", Description: "会议时长（秒）"},
				"meeting_type":            {Type: "integer", Description: "会议类型（0=视频, 1=语音）"},
				"invitees": {Type: "object", Properties: map[string]SchemaRef{
					"userid":     {Type: "array", Items: strSchema()},
					"department": {Type: "array", Items: strSchema()},
				}},
			},
			Required: []string{"title", "meeting_start_datetime", "meeting_duration"},
		},
		"MeetingInfo": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"meetingid":              {Type: "string"},
				"title":                  {Type: "string"},
				"status":                 {Type: "integer"},
				"meeting_start_datetime": {Type: "string", Format: "date-time"},
				"meeting_duration":       {Type: "integer"},
				"creator":                {Type: "string"},
				"meeting_link":           {Type: "string"},
			},
		},
		"MeetingList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data":    {Type: "object", Properties: map[string]SchemaRef{
					"meetings": {Type: "array", Items: schemaRef("#/components/schemas/MeetingInfo")},
				}},
			},
		},

		// Message
		"SendTextRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"receiver_ids": {Type: "array", Items: strSchema(), Description: "接收人UserID列表"},
				"content":      {Type: "string", Description: "消息内容"},
			},
			Required: []string{"receiver_ids", "content"},
		},
		"SendImageRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"receiver_ids": {Type: "array", Items: strSchema()},
				"image_url":    {Type: "string", Description: "图片URL"},
				"media_id":     {Type: "string", Description: "素材ID"},
			},
		},
		"SendFileRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"receiver_ids": {Type: "array", Items: strSchema()},
				"file_url":     {Type: "string", Description: "文件URL"},
				"media_id":     {Type: "string", Description: "素材ID"},
			},
		},
		"SendCardRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"receiver_ids":  {Type: "array", Items: strSchema()},
				"card_content":  {Type: "object", Description: "卡片内容"},
			},
		},
		"SendResult": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"invalid_user_ids":  {Type: "array", Items: strSchema()},
				"invalid_party_ids": {Type: "array", Items: strSchema()},
				"invalid_tag_ids":   {Type: "array", Items: strSchema()},
				"failed_user_ids":   {Type: "array", Items: strSchema()},
			},
		},

		// Document
		"CreateDocumentRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"title":    {Type: "string", Description: "文档标题"},
				"type":     {Type: "string", Description: "文档类型 (doc/sheet/mindnote)"},
				"space_id": {Type: "string", Description: "知识空间ID"},
			},
			Required: []string{"title"},
		},
		"DocumentInfo": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"doc_id":   {Type: "string"},
				"title":    {Type: "string"},
				"type":     {Type: "string"},
				"url":      {Type: "string"},
				"owner":    {Type: "string"},
			},
		},
		"DocumentList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data":    {Type: "object", Properties: map[string]SchemaRef{
					"documents": {Type: "array", Items: schemaRef("#/components/schemas/DocumentInfo")},
				}},
			},
		},
		"RenameDocumentRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"name": {Type: "string", Description: "新名称"},
			},
			Required: []string{"name"},
		},
		"EditContentRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"content": {Type: "string", Description: "文档内容（Markdown）"},
			},
			Required: []string{"content"},
		},
		"ShareDocumentRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"users":   {Type: "array", Items: strSchema(), Description: "分享用户ID列表"},
				"comment": {Type: "string", Description: "分享备注"},
			},
		},

		// Sheet
		"CreateSheetRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"title": {Type: "string", Description: "表格标题"},
			},
			Required: []string{"title"},
		},

		// Contact
		"ContactUser": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"userid":     {Type: "string"},
				"name":       {Type: "string"},
				"alias":      {Type: "string"},
				"mobile":     {Type: "string"},
				"email":      {Type: "string"},
				"department": {Type: "array", Items: intSchema()},
				"position":   {Type: "string"},
				"gender":     {Type: "integer"},
				"status":     {Type: "integer"},
				"avatar":     {Type: "string"},
			},
		},
		"ContactUserList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data":    {Type: "object", Properties: map[string]SchemaRef{
					"users": {Type: "array", Items: schemaRef("#/components/schemas/ContactUser")},
					"count": {Type: "integer"},
				}},
			},
		},

		// Todo
		"CreateTodoRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"content":     {Type: "string", Description: "待办内容"},
				"assignees":   {Type: "array", Items: strSchema(), Description: "指派人"},
				"remind_time": {Type: "string", Format: "date-time", Description: "提醒时间"},
			},
			Required: []string{"content"},
		},
		"UpdateTodoRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"content": {Type: "string", Description: "待办内容"},
				"status":  {Type: "integer", Description: "状态 (0=完成, 1=进行中)"},
			},
		},
		"ChangeTodoStatusRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"status": {Type: "integer", Description: "状态 (0=完成, 1=进行中)"},
			},
			Required: []string{"status"},
		},
		"TodoCreateResult": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data":    {Type: "object", Properties: map[string]SchemaRef{
					"todo_id": {Type: "string"},
				}},
			},
		},
		"TodoDetail": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"todo_id":     {Type: "string"},
				"todo_status": {Type: "integer"},
				"user_status": {Type: "integer"},
				"creator_id":  {Type: "string"},
				"content":     {Type: "string"},
				"assignees":   {Type: "array", Items: strSchema()},
				"remind_time": {Type: "string", Format: "date-time"},
				"create_time": {Type: "string", Format: "date-time"},
				"update_time": {Type: "string", Format: "date-time"},
			},
		},
		"TodoList": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"code":    {Type: "integer", Example: 0},
				"message": {Type: "string", Example: "ok"},
				"data":    {Type: "object", Properties: map[string]SchemaRef{
					"index_list": {Type: "array", Items: schemaRef("#/components/schemas/TodoDetail")},
					"has_more":   {Type: "boolean"},
					"next_cursor": {Type: "string"},
				}},
			},
		},

		// MCP
		"MCPRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"jsonrpc": {Type: "string", Description: "JSON-RPC 版本，固定为 2.0"},
				"id":      {Type: "integer", Description: "请求ID"},
				"method":  {Type: "string", Description: "方法名"},
				"params":  {Type: "object", Description: "方法参数"},
			},
			Required: []string{"jsonrpc", "method"},
		},
		"MCPResponse": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"jsonrpc": {Type: "string"},
				"id":      {Type: "integer"},
				"result":  {Type: "object"},
				"error": {Type: "object", Properties: map[string]SchemaRef{
					"code":    {Type: "integer"},
					"message": {Type: "string"},
				}},
			},
		},

		// Admin
		"LoginRequest": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"username": {Type: "string"},
				"password": {Type: "string"},
			},
			Required: []string{"username", "password"},
		},
		"LoginResponse": {
			Type: "object",
			Properties: map[string]SchemaRef{
				"token":      {Type: "string"},
				"expires_at": {Type: "string", Format: "date-time"},
			},
		},
	}
}

// Helper functions for building spec

func strSchema() *SchemaRef {
	return &SchemaRef{Type: "string"}
}

func intSchema() *SchemaRef {
	return &SchemaRef{Type: "integer"}
}

func arrSchema(items *SchemaRef) *SchemaRef {
	return &SchemaRef{Type: "array", Items: items}
}

func schemaRef(ref string) *SchemaRef {
	return &SchemaRef{Ref: ref}
}

func pathParam(name, desc string) Parameter {
	return Parameter{Name: name, In: "path", Description: desc, Required: true, Schema: strSchema()}
}

func pathParams(ps ...Parameter) []Parameter {
	return ps
}

func jsonBody(desc string, required bool, schemaRef string) *RequestBody {
	return &RequestBody{
		Description: desc,
		Required:    required,
		Content: map[string]MediaType{
			"application/json": {
				Schema: &SchemaRef{Ref: schemaRef},
			},
		},
	}
}

func response(desc, schemaRef string) Response {
	return Response{
		Description: desc,
		Content: map[string]MediaType{
			"application/json": {
				Schema: &SchemaRef{Ref: schemaRef},
			},
		},
	}
}

func errResponse(desc string) Response {
	return Response{Description: desc}
}
