package openapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetSpec(t *testing.T) {
	spec := GetSpec()

	if spec.OpenAPI != "3.0.3" {
		t.Errorf("expected OpenAPI version 3.0.3, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != "WeCom Gateway API" {
		t.Errorf("expected title 'WeCom Gateway API', got %s", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", spec.Info.Version)
	}

	if len(spec.Servers) == 0 {
		t.Error("expected at least one server")
	}

	if len(spec.Tags) == 0 {
		t.Error("expected tags")
	}
}

func TestGetSpecJSON(t *testing.T) {
	data, err := GetSpecJSON()
	if err != nil {
		t.Fatalf("GetSpecJSON() returned error: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("GetSpecJSON() returned empty data")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("GetSpecJSON() returned invalid JSON: %v", err)
	}

	if result["openapi"] != "3.0.3" {
		t.Errorf("expected openapi 3.0.3, got %v", result["openapi"])
	}
}

func TestSpecPaths_ContainsRequiredEndpoints(t *testing.T) {
	spec := GetSpec()
	requiredPaths := []string{
		"/v1/schedules",
		"/v1/meeting-rooms",
		"/v1/meetings",
		"/v1/messages/text",
		"/v1/messages/markdown",
		"/v1/docs",
		"/v1/sheets",
		"/v1/contacts/users",
		"/v1/todos",
		"/v1/admin/login",
		"/v1/admin/api-keys",
		"/mcp",
	}

	for _, path := range requiredPaths {
		if _, exists := spec.Paths[path]; !exists {
			t.Errorf("missing required path: %s", path)
		}
	}
}

func TestSpecPaths_ScheduleOperations(t *testing.T) {
	spec := GetSpec()

	// Check /v1/schedules has POST and GET
	schedulesPath, exists := spec.Paths["/v1/schedules"]
	if !exists {
		t.Fatal("/v1/schedules path not found")
	}

	if _, exists := schedulesPath["post"]; !exists {
		t.Error("missing POST on /v1/schedules")
	}
	if _, exists := schedulesPath["get"]; !exists {
		t.Error("missing GET on /v1/schedules")
	}

	// Check /v1/schedules/{id} has GET, PATCH, DELETE
	scheduleIDPath, exists := spec.Paths["/v1/schedules/{id}"]
	if !exists {
		t.Fatal("/v1/schedules/{id} path not found")
	}

	for _, method := range []string{"get", "patch", "delete"} {
		if _, exists := scheduleIDPath[method]; !exists {
			t.Errorf("missing %s on /v1/schedules/{id}", method)
		}
	}
}

func TestSpecPaths_TodoOperations(t *testing.T) {
	spec := GetSpec()

	todosPath, exists := spec.Paths["/v1/todos"]
	if !exists {
		t.Fatal("/v1/todos path not found")
	}

	for _, method := range []string{"get", "post"} {
		if _, exists := todosPath[method]; !exists {
			t.Errorf("missing %s on /v1/todos", method)
		}
	}

	todoIDPath, exists := spec.Paths["/v1/todos/{id}"]
	if !exists {
		t.Fatal("/v1/todos/{id} path not found")
	}

	for _, method := range []string{"get", "put", "delete"} {
		if _, exists := todoIDPath[method]; !exists {
			t.Errorf("missing %s on /v1/todos/{id}", method)
		}
	}

	if _, exists := spec.Paths["/v1/todos/{id}/status"]; !exists {
		t.Error("missing /v1/todos/{id}/status path")
	}
}

func TestSpecComponents_SecuritySchemes(t *testing.T) {
	spec := GetSpec()

	if spec.Components.SecuritySchemes == nil {
		t.Fatal("missing security schemes")
	}

	bearerAuth, exists := spec.Components.SecuritySchemes["BearerAuth"]
	if !exists {
		t.Fatal("missing BearerAuth security scheme")
	}

	if bearerAuth.Type != "http" {
		t.Errorf("expected type 'http', got %s", bearerAuth.Type)
	}

	if bearerAuth.Scheme != "bearer" {
		t.Errorf("expected scheme 'bearer', got %s", bearerAuth.Scheme)
	}
}

func TestSpecComponents_Schemas(t *testing.T) {
	spec := GetSpec()

	requiredSchemas := []string{
		"ApiResponse",
		"Schedule",
		"CreateScheduleRequest",
		"MeetingRoom",
		"SendTextRequest",
		"SendResult",
		"ContactUser",
		"CreateTodoRequest",
		"DocumentInfo",
		"MCPRequest",
		"MCPResponse",
	}

	for _, name := range requiredSchemas {
		if _, exists := spec.Components.Schemas[name]; !exists {
			t.Errorf("missing required schema: %s", name)
		}
	}
}

func TestSpecOperations_HaveSecurity(t *testing.T) {
	spec := GetSpec()

	// Most endpoints should have security
	securedPaths := []string{
		"/v1/schedules",
		"/v1/meeting-rooms",
		"/v1/meetings",
		"/v1/messages/text",
		"/v1/contacts/users",
		"/v1/todos",
	}

	for _, path := range securedPaths {
		pathItem := spec.Paths[path]
		for method, op := range pathItem {
			if len(op.Security) == 0 {
				t.Errorf("%s %s missing security", strings.ToUpper(method), path)
			}
		}
	}
}

func TestSpec_Tags(t *testing.T) {
	spec := GetSpec()

	expectedTags := map[string]bool{
		"schedules":     false,
		"meeting-rooms": false,
		"meetings":      false,
		"messages":      false,
		"documents":     false,
		"sheets":        false,
		"contacts":      false,
		"todos":         false,
		"admin":         false,
		"mcp":           false,
	}

	for _, tag := range spec.Tags {
		if _, exists := expectedTags[tag.Name]; exists {
			expectedTags[tag.Name] = true
		}
	}

	for name, found := range expectedTags {
		if !found {
			t.Errorf("missing tag: %s", name)
		}
	}
}

func TestSpec_MCPPath(t *testing.T) {
	spec := GetSpec()

	mcpPath, exists := spec.Paths["/mcp"]
	if !exists {
		t.Fatal("/mcp path not found")
	}

	if _, exists := mcpPath["get"]; !exists {
		t.Error("missing GET on /mcp (SSE)")
	}

	if _, exists := mcpPath["post"]; !exists {
		t.Error("missing POST on /mcp (JSON-RPC)")
	}

	postOp := mcpPath["post"]
	if len(postOp.Tags) == 0 || postOp.Tags[0] != "mcp" {
		t.Error("POST /mcp should have 'mcp' tag")
	}
}
