// Package grpcadapter provides a minimal IdentityService stub (no generated protobuf yet).
package grpcadapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app"
	"github.com/nexora/identity-service/internal/domain/policy"
)

// Server implements CheckPermission and ValidateSession without gRPC codegen.
type Server struct {
	Deps *app.Deps
}

// CheckPermissionRequest mirrors proto CheckPermissionRequest.
type CheckPermissionRequest struct {
	PrincipalID string
	Permission  string
	TenantID    string
	CityID      string
	WarehouseID string
}

// CheckPermissionResponse mirrors proto CheckPermissionResponse.
type CheckPermissionResponse struct {
	Allowed bool
	Reason  string
}

// ValidateSessionRequest mirrors proto ValidateSessionRequest.
type ValidateSessionRequest struct {
	SessionID   string
	AccessToken string
}

// ValidateSessionResponse mirrors proto ValidateSessionResponse.
type ValidateSessionResponse struct {
	Valid       bool
	PrincipalID string
	TenantID    string
	Roles       []string
	Reason      string
}

// CheckPermission is the PDP stub wired to app.Deps.
func (s *Server) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (*CheckPermissionResponse, error) {
	if s.Deps == nil {
		return &CheckPermissionResponse{Allowed: false, Reason: "deps not configured"}, nil
	}
	pid, err := uuid.Parse(req.PrincipalID)
	if err != nil {
		return &CheckPermissionResponse{Allowed: false, Reason: "invalid principal_id"}, nil
	}
	res, err := s.Deps.CheckPermission(ctx, app.CheckPermissionInput{
		PrincipalID: pid,
		Permission:  req.Permission,
		Attrs:       policy.Attributes{},
	})
	if err != nil {
		return nil, err
	}
	reason := ""
	if len(res.Reasons) > 0 {
		reason = res.Reasons[0]
	}
	return &CheckPermissionResponse{Allowed: res.Allow, Reason: reason}, nil
}

// ValidateSession validates a session id via repository.
func (s *Server) ValidateSession(ctx context.Context, req *ValidateSessionRequest) (*ValidateSessionResponse, error) {
	if s.Deps == nil {
		return &ValidateSessionResponse{Valid: false, Reason: "deps not configured"}, nil
	}
	if req.AccessToken != "" && s.Deps.JWTKeys != nil {
		claims, err := s.Deps.JWTKeys.ParseAndValidate(req.AccessToken, s.Deps.Issuer, s.Deps.Audience)
		if err != nil {
			return &ValidateSessionResponse{Valid: false, Reason: err.Error()}, nil
		}
		return &ValidateSessionResponse{
			Valid: true, PrincipalID: claims.Subject, TenantID: claims.Tenant, Roles: claims.Roles,
		}, nil
	}
	if req.SessionID == "" {
		return &ValidateSessionResponse{Valid: false, Reason: "missing session or token"}, nil
	}
	sid, err := uuid.Parse(req.SessionID)
	if err != nil {
		return &ValidateSessionResponse{Valid: false, Reason: "invalid session_id"}, nil
	}
	sess, err := s.Deps.Sessions.GetByID(ctx, sid)
	if err != nil {
		return &ValidateSessionResponse{Valid: false, Reason: err.Error()}, nil
	}
	if !sess.IsUsable(time.Now().UTC()) {
		return &ValidateSessionResponse{Valid: false, Reason: "session not usable"}, nil
	}
	return &ValidateSessionResponse{
		Valid: true, PrincipalID: sess.PrincipalID.String(), TenantID: sess.TenantID.String(),
	}, nil
}
