package grpcserver

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/store"
)

func TestRecoveryInterceptor(t *testing.T) {
	interceptor := RecoveryInterceptor()

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "response", nil
	}

	_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("handler was not called")
	}
}

func TestRecoveryInterceptor_Panic(t *testing.T) {
	interceptor := RecoveryInterceptor()

	panicHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("test panic")
	}

	_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{}, panicHandler)
	if err == nil {
		t.Error("expected error from panic")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected grpc status error")
	}
	if st.Code() != codes.Internal {
		t.Errorf("expected code Internal, got %v", st.Code())
	}
}

func TestLoggingInterceptor(t *testing.T) {
	interceptor := LoggingInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoggingInterceptor_Error(t *testing.T) {
	interceptor := LoggingInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	_, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}, handler)
	if err == nil {
		t.Error("expected error")
	}
}

func TestAuthInterceptor_HealthCheck(t *testing.T) {
	authenticator := &auth.APIKeyAuthenticator{}
	interceptor := AuthInterceptor(authenticator)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	// Health check should skip auth
	info := &grpc.UnaryServerInfo{FullMethod: "/wecom.gateway.WeComGateway/Health"}
	_, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuthInterceptor_MissingAPIKey(t *testing.T) {
	authenticator := &auth.APIKeyAuthenticator{}
	interceptor := AuthInterceptor(authenticator)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	ctx := context.Background()

	_, err := interceptor(ctx, "request", info, handler)
	if err == nil {
		t.Error("expected error for missing API key")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Error("expected grpc status error")
	}
	if st.Code() != codes.Unauthenticated {
		t.Errorf("expected code Unauthenticated, got %v", st.Code())
	}
}

func TestAuditInterceptor(t *testing.T) {
	// Use a real Logger with in-memory SQLite so Log() doesn't panic on nil db
	db, _ := store.NewSQLite(":memory:")
	defer db.Close()
	logger := audit.NewLogger(db)
	interceptor := AuditInterceptor(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuditInterceptor_WithError(t *testing.T) {
	db, _ := store.NewSQLite(":memory:")
	defer db.Close()
	logger := audit.NewLogger(db)
	interceptor := AuditInterceptor(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	_, err := interceptor(context.Background(), "request", info, handler)
	if err == nil {
		t.Error("expected error")
	}
}

func TestAuditInterceptor_WithDuration(t *testing.T) {
	db, _ := store.NewSQLite(":memory:")
	defer db.Close()
	logger := audit.NewLogger(db)
	interceptor := AuditInterceptor(logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	start := time.Now()

	_, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	duration := time.Since(start)
	if duration < 10*time.Millisecond {
		t.Error("expected handler to take at least 10ms")
	}
}

func TestRateLimitInterceptor_Allowed(t *testing.T) {
	limiter := NewRateLimiter()
	interceptor := RateLimitInterceptor(limiter, 100)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	ctx := context.Background()

	// Add auth context
	authCtx := &auth.AuthContext{KeyID: "test-key"}
	ctx = context.WithValue(ctx, "auth_context", authCtx)

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRateLimitInterceptor_NoAuthContext(t *testing.T) {
	limiter := NewRateLimiter()
	interceptor := RateLimitInterceptor(limiter, 100)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	// No auth context - should pass through
	_, err := interceptor(context.Background(), "request", info, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter()

	// Mock implementation always returns true
	if !limiter.Allow("test-key", 100) {
		t.Error("expected Allow to return true")
	}

	// Test with different limits
	if !limiter.Allow("test-key", 10) {
		t.Error("expected Allow to return true with limit 10")
	}

	if !limiter.Allow("test-key", 1000) {
		t.Error("expected Allow to return true with limit 1000")
	}
}

func TestInterceptor_Chaining(t *testing.T) {
	// Test that multiple interceptors can be chained
	callOrder := []string{}

	interceptor1 := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		callOrder = append(callOrder, "interceptor1")
		return handler(ctx, req)
	}

	interceptor2 := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		callOrder = append(callOrder, "interceptor2")
		return handler(ctx, req)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		callOrder = append(callOrder, "handler")
		return "response", nil
	}

	// Chain interceptors
	chain := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return interceptor1(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
			return interceptor2(ctx, req, info, handler)
		})
	}

	_, err := chain(context.Background(), "request", &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(callOrder) != 3 {
		t.Errorf("expected 3 calls, got %d", len(callOrder))
	}
	if callOrder[0] != "interceptor1" {
		t.Errorf("expected first call to be interceptor1, got %s", callOrder[0])
	}
	if callOrder[1] != "interceptor2" {
		t.Errorf("expected second call to be interceptor2, got %s", callOrder[1])
	}
	if callOrder[2] != "handler" {
		t.Errorf("expected third call to be handler, got %s", callOrder[2])
	}
}
