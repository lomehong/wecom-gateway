package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"wecom-gateway/internal/apikey"
	"wecom-gateway/internal/audit"
	"wecom-gateway/internal/auth"
	"wecom-gateway/internal/config"
	"wecom-gateway/internal/meeting"
	"wecom-gateway/internal/message"
	"wecom-gateway/internal/schedule"
	"wecom-gateway/internal/wecom"
	proto "wecom-gateway/api/proto"
)

// Server implements the gRPC server
type Server struct {
	proto.UnimplementedWeComGatewayServer
	proto.UnimplementedWeComGatewayAdminServer

	cfg         *config.Config
	wecomClient wecom.Client
	scheduleSvc *schedule.Service
	meetingSvc  *meeting.Service
	messageSvc  *message.Service
	apiKeySvc   *apikey.Service
	auditLogger *audit.Logger
	auth        *auth.APIKeyAuthenticator
	rateLimiter *RateLimiter
}

// RateLimiter is a simple rate limiter for gRPC
type RateLimiter struct {
	// Implementation would go here
	// For now, it's a placeholder
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(apiKey string, requestsPerMinute int) bool {
	// Placeholder implementation
	// In production, use token bucket or similar algorithm
	return true
}

// NewServer creates a new gRPC server
func NewServer(
	cfg *config.Config,
	wecomClient wecom.Client,
	scheduleSvc *schedule.Service,
	meetingSvc *meeting.Service,
	messageSvc *message.Service,
	apiKeySvc *apikey.Service,
	auditLogger *audit.Logger,
	auth *auth.APIKeyAuthenticator,
) *Server {
	return &Server{
		cfg:         cfg,
		wecomClient: wecomClient,
		scheduleSvc: scheduleSvc,
		meetingSvc:  meetingSvc,
		messageSvc:  messageSvc,
		apiKeySvc:   apiKeySvc,
		auditLogger: auditLogger,
		auth:        auth,
		rateLimiter: NewRateLimiter(),
	}
}

// Start starts the gRPC server
func (s *Server) Start(listenAddr string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with interceptors
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RecoveryInterceptor(),
			LoggingInterceptor(),
			AuthInterceptor(s.auth),
			RateLimitInterceptor(s.rateLimiter, s.cfg.Auth.RateLimit),
			AuditInterceptor(s.auditLogger),
		),
	)

	// Register services
	proto.RegisterWeComGatewayServer(server, s)
	proto.RegisterWeComGatewayAdminServer(server, s)

	// Start server
	log.Printf("Starting gRPC server on %s", listenAddr)
	return server.Serve(lis)
}

// Schedule operations

func (s *Server) CreateSchedule(ctx context.Context, req *proto.CreateScheduleRequest) (*proto.Schedule, error) {
	// Extract corp and app from auth context
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.ScheduleParams{
		Organizer:        req.Organizer,
		Summary:          req.Summary,
		Description:      req.Description,
		StartTime:        time.Unix(req.StartTime, 0),
		EndTime:          time.Unix(req.EndTime, 0),
		Attendees:        req.Attendees,
		Location:         req.Location,
		RemindBeforeMin:  int(req.RemindBeforeMinutes),
	}

	schedule, err := s.wecomClient.CreateSchedule(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.Schedule{
		ScheduleId:  schedule.ScheduleID,
		Organizer:   schedule.Organizer,
		Summary:     schedule.Summary,
		Description: schedule.Description,
		StartTime:   schedule.StartTime.Unix(),
		EndTime:     schedule.EndTime.Unix(),
		Attendees:   schedule.Attendees,
		Location:    schedule.Location,
		CalId:       schedule.CalID,
	}, nil
}

func (s *Server) GetSchedules(ctx context.Context, req *proto.GetSchedulesRequest) (*proto.ScheduleList, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	schedules, err := s.wecomClient.GetSchedules(
		ctx,
		authCtx.CorpName,
		authCtx.AppName,
		req.Userid,
		time.Unix(req.StartTime, 0),
		time.Unix(req.EndTime, 0),
		int(req.Limit),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoSchedules := make([]*proto.Schedule, len(schedules))
	for i, sched := range schedules {
		protoSchedules[i] = &proto.Schedule{
			ScheduleId:  sched.ScheduleID,
			Organizer:   sched.Organizer,
			Summary:     sched.Summary,
			Description: sched.Description,
			StartTime:   sched.StartTime.Unix(),
			EndTime:     sched.EndTime.Unix(),
			Attendees:   sched.Attendees,
			Location:    sched.Location,
			CalId:       sched.CalID,
		}
	}

	return &proto.ScheduleList{Schedules: protoSchedules}, nil
}

func (s *Server) UpdateSchedule(ctx context.Context, req *proto.UpdateScheduleRequest) (*proto.Schedule, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.ScheduleParams{
		Summary:         req.Summary,
		Description:     req.Description,
		StartTime:       time.Unix(req.StartTime, 0),
		EndTime:         time.Unix(req.EndTime, 0),
		Attendees:       req.Attendees,
		Location:        req.Location,
		RemindBeforeMin: int(req.RemindBeforeMinutes),
	}

	if err := s.wecomClient.UpdateSchedule(ctx, authCtx.CorpName, authCtx.AppName, req.ScheduleId, params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return updated schedule (would need to fetch it, but for now return success)
	return &proto.Schedule{ScheduleId: req.ScheduleId}, nil
}

func (s *Server) DeleteSchedule(ctx context.Context, req *proto.DeleteScheduleRequest) (*proto.DeleteResponse, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	if err := s.wecomClient.DeleteSchedule(ctx, authCtx.CorpName, authCtx.AppName, req.ScheduleId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.DeleteResponse{Message: "Schedule deleted successfully"}, nil
}

// Meeting room operations

func (s *Server) ListMeetingRooms(ctx context.Context, req *proto.ListMeetingRoomsRequest) (*proto.MeetingRoomList, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	opts := &wecom.RoomQueryOptions{
		City:     req.City,
		Building: req.Building,
		Floor:    req.Floor,
		Capacity: int(req.Capacity),
		Equipment: req.Equipment,
		Limit:    int(req.Limit),
		Cursor:   req.Cursor,
	}

	rooms, nextCursor, err := s.wecomClient.ListMeetingRooms(ctx, authCtx.CorpName, authCtx.AppName, opts)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoRooms := make([]*proto.MeetingRoom, len(rooms))
	for i, room := range rooms {
		protoRooms[i] = &proto.MeetingRoom{
			MbookingId:  room.MBookingID,
			Name:        room.Name,
			Capacity:    int32(room.Capacity),
			City:        room.City,
			Building:    room.Building,
			Floor:       room.Floor,
			Equipment:   room.Equipment,
			Description: room.Description,
		}
	}

	return &proto.MeetingRoomList{Rooms: protoRooms, NextCursor: nextCursor}, nil
}

func (s *Server) GetRoomAvailability(ctx context.Context, req *proto.AvailabilityRequest) (*proto.TimeSlotList, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	slots, err := s.wecomClient.GetRoomAvailability(
		ctx,
		authCtx.CorpName,
		authCtx.AppName,
		req.MeetingroomId,
		time.Unix(req.StartTime, 0),
		time.Unix(req.EndTime, 0),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoSlots := make([]*proto.TimeSlot, len(slots))
	for i, slot := range slots {
		protoSlots[i] = &proto.TimeSlot{
			StartTime: slot.StartTime.Unix(),
			EndTime:   slot.EndTime.Unix(),
		}
	}

	return &proto.TimeSlotList{Slots: protoSlots}, nil
}

func (s *Server) BookMeetingRoom(ctx context.Context, req *proto.BookMeetingRoomRequest) (*proto.BookingResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.BookingParams{
		MeetingRoomID: req.MeetingroomId,
		Subject:       req.Subject,
		StartTime:     time.Unix(req.StartTime, 0),
		EndTime:       time.Unix(req.EndTime, 0),
		Booker:        req.Booker,
		Attendees:     req.Attendees,
	}

	result, err := s.wecomClient.BookMeetingRoom(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.BookingResult{
		BookingId:   result.BookingID,
		ScheduleId:  result.ScheduleID,
		StartTime:   result.StartTime.Unix(),
		EndTime:     result.EndTime.Unix(),
	}, nil
}

// Message operations

func (s *Server) SendText(ctx context.Context, req *proto.TextMessageRequest) (*proto.SendResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.MessageParams{
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIds,
		Content:      req.Content,
		Safe:         req.Safe,
	}

	result, err := s.wecomClient.SendText(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertSendResult(result), nil
}

func (s *Server) SendMarkdown(ctx context.Context, req *proto.MarkdownMessageRequest) (*proto.SendResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.MessageParams{
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIds,
		Content:      req.Content,
	}

	result, err := s.wecomClient.SendMarkdown(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertSendResult(result), nil
}

func (s *Server) SendImage(ctx context.Context, req *proto.ImageMessageRequest) (*proto.SendResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.ImageMessageParams{
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIds,
		MediaID:      req.MediaId,
	}

	result, err := s.wecomClient.SendImage(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertSendResult(result), nil
}

func (s *Server) SendFile(ctx context.Context, req *proto.FileMessageRequest) (*proto.SendResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	params := &wecom.FileMessageParams{
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIds,
		MediaID:      req.MediaId,
	}

	result, err := s.wecomClient.SendFile(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertSendResult(result), nil
}

func (s *Server) SendCard(ctx context.Context, req *proto.CardMessageRequest) (*proto.SendResult, error) {
	authCtx, ok := ctx.Value("auth_context").(*auth.AuthContext)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing auth context")
	}

	// Convert map[string]string to map[string]interface{}
	cardContent := make(map[string]interface{}, len(req.CardContent))
	for k, v := range req.CardContent {
		cardContent[k] = v
	}

	params := &wecom.CardMessageParams{
		ReceiverType: req.ReceiverType,
		ReceiverIDs:  req.ReceiverIds,
		CardContent:  cardContent,
	}

	result, err := s.wecomClient.SendCard(ctx, authCtx.CorpName, authCtx.AppName, params)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertSendResult(result), nil
}

// Admin operations

func (s *Server) CreateAPIKey(ctx context.Context, req *proto.CreateAPIKeyRequest) (*proto.APIKeyInfo, error) {
	// This would call the API key service
	// For now, return a placeholder
	return &proto.APIKeyInfo{
		Id:          "test-id",
		Name:        req.Name,
		Permissions: req.Permissions,
		CorpName:    req.CorpName,
		AppName:     req.AppName,
		ExpiresAt:   time.Now().Add(24 * time.Hour * time.Duration(req.ExpiresDays)).Unix(),
		Disabled:    false,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}, nil
}

func (s *Server) ListAPIKeys(ctx context.Context, req *proto.ListAPIKeysRequest) (*proto.APIKeyList, error) {
	// Placeholder implementation
	return &proto.APIKeyList{
		Keys:       []*proto.APIKeyInfo{},
		NextCursor: "",
	}, nil
}

func (s *Server) DeleteAPIKey(ctx context.Context, req *proto.DeleteAPIKeyRequest) (*proto.DeleteResponse, error) {
	// Placeholder implementation
	return &proto.DeleteResponse{Message: "API key deleted successfully"}, nil
}

func (s *Server) QueryAuditLogs(ctx context.Context, req *proto.QueryAuditLogsRequest) (*proto.AuditLogList, error) {
	// Placeholder implementation
	return &proto.AuditLogList{
		Logs:       []*proto.AuditLog{},
		NextCursor: "",
	}, nil
}

func (s *Server) GetDashboardStats(ctx context.Context, req *proto.DashboardStatsRequest) (*proto.DashboardStats, error) {
	// Placeholder implementation
	return &proto.DashboardStats{
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		TotalApiKeys:   0,
		ActiveApiKeys:  0,
		TotalCorps:     0,
		TotalApps:      0,
		TotalRequests:  0,
		ErrorRequests:  0,
	}, nil
}

// Helper functions

func convertSendResult(result *wecom.SendResult) *proto.SendResult {
	return &proto.SendResult{
		InvalidUserIds:  result.InvalidUserIDs,
		InvalidPartyIds: result.InvalidPartyIDs,
		InvalidTagIds:   result.InvalidTagIDs,
		UnquotedUserIds: result.UnquotedUserIDs,
		FailedUserIds:   result.FailedUserIDs,
		FailedPartyIds:  result.FailedPartyIDs,
		FailedTagIds:    result.FailedTagIDs,
	}
}
