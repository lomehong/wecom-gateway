package grpcserver

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/audit"
)

// AuthInterceptor creates an authentication interceptor
func AuthInterceptor(authenticator *auth.APIKeyAuthenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip auth for health check
		if info.FullMethod == "/wecom.gateway.WeComGateway/Health" {
			return handler(ctx, req)
		}

		// Extract API key from metadata
		var apiKey string
		if md, ok := ctx.Value("auth_metadata").(map[string][]string); ok {
			if keys, ok := md["authorization"]; ok && len(keys) > 0 {
				apiKey = keys[0]
			}
		}

		if apiKey == "" {
			return nil, status.Error(codes.Unauthenticated, "missing API key")
		}

		// Validate API key
		authCtx, err := authenticator.Authenticate(ctx, apiKey)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid API key")
		}

		// Add auth context to context
		ctx = context.WithValue(ctx, "auth_context", authCtx)

		return handler(ctx, req)
	}
}

// AuditInterceptor creates an audit logging interceptor
func AuditInterceptor(logger *audit.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Extract auth context if available
		var apiKeyID, apiKeyName *string
		if authCtx, exists := ctx.Value("auth_context").(*auth.AuthContext); exists {
			apiKeyName = &authCtx.KeyName
		}

		// Process request
		resp, err := handler(ctx, req)

		// Log after request completes
		duration := time.Since(start)
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			}
		}

		logger.Log(ctx, &audit.LogEntry{
			Timestamp:  start,
			Protocol:   "grpc",
			APIKeyID:   apiKeyID,
			APIKeyName: apiKeyName,
			Method:     info.FullMethod,
			Path:       info.FullMethod,
			StatusCode: int(statusCode),
			DurationMs: int(duration.Milliseconds()),
		})

		return resp, err
	}
}

// RateLimitInterceptor creates a rate limiting interceptor
func RateLimitInterceptor(limiter *RateLimiter, requestsPerMinute int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Extract API key from context
		var apiKey string
		if authCtx, exists := ctx.Value("auth_context").(*auth.AuthContext); exists {
			apiKey = authCtx.KeyID
		}

		if apiKey == "" {
			return handler(ctx, req)
		}

		// Check rate limit
		if !limiter.Allow(apiKey, requestsPerMinute) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// RecoveryInterceptor creates a recovery interceptor
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered in gRPC call %s: %v", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// LoggingInterceptor creates a logging interceptor for debugging
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		log.Printf("gRPC call started: %s", info.FullMethod)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		if err != nil {
			log.Printf("gRPC call failed: %s (duration: %v, error: %v)", info.FullMethod, duration, err)
		} else {
			log.Printf("gRPC call completed: %s (duration: %v)", info.FullMethod, duration)
		}

		return resp, err
	}
}
