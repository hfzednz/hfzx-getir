// Package grpcadapter provides ProfileService stubs (no generated protobuf yet).
package grpcadapter

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app"
	"github.com/nexora/customer-profile-service/internal/domain"
)

// Server implements GetProfile, GetPreferences, ListAddresses, CheckConsent, GetPersonalization.
type Server struct {
	Deps *app.Deps
}

// ---------------------------------------------------------------------------
// Request / response mirrors of proto/profile/v1/profile.proto
// ---------------------------------------------------------------------------

type GetProfileRequest struct {
	ProfileID string
}

type GetProfileResponse struct {
	Profile *ProfileMsg
}

type GetPreferencesRequest struct {
	ProfileID string
}

type GetPreferencesResponse struct {
	Preferences *PreferencesMsg
}

type ListAddressesRequest struct {
	ProfileID string
}

type ListAddressesResponse struct {
	Addresses []*AddressMsg
}

type CheckConsentRequest struct {
	ProfileID string
	Channel   string
}

type CheckConsentResponse struct {
	Granted bool
}

type GetPersonalizationRequest struct {
	ProfileID string
}

type GetPersonalizationResponse struct {
	Personalization *PersonalizationMsg
}

type ProfileMsg struct {
	Id          string
	PrincipalId string
	TenantId    string
	DisplayName string
	FullName    string
	Language    string
	CountryCode string
	City        string
	Status      string
	AvatarUrl   string
}

type PreferencesMsg struct {
	ProfileId string
	Theme     string
	Language  string
}

type AddressMsg struct {
	Id        string
	ProfileId string
	Label     string
	Line1     string
	Lat       float64
	Lng       float64
	IsDefault bool
}

type PersonalizationMsg struct {
	ProfileId string
}

// GetProfile returns a profile by id.
func (s *Server) GetProfile(ctx context.Context, req *GetProfileRequest) (*GetProfileResponse, error) {
	if s.Deps == nil {
		return nil, domain.ErrInvariant
	}
	id, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	p, err := s.Deps.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetProfileResponse{Profile: toProfileMsg(p)}, nil
}

// GetPreferences returns preferences for a profile.
func (s *Server) GetPreferences(ctx context.Context, req *GetPreferencesRequest) (*GetPreferencesResponse, error) {
	if s.Deps == nil {
		return nil, domain.ErrInvariant
	}
	id, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	p, err := s.Deps.GetPreferences(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetPreferencesResponse{Preferences: &PreferencesMsg{
		ProfileId: p.ProfileID.String(), Theme: p.Theme, Language: p.Language,
	}}, nil
}

// ListAddresses returns addresses for a profile.
func (s *Server) ListAddresses(ctx context.Context, req *ListAddressesRequest) (*ListAddressesResponse, error) {
	if s.Deps == nil {
		return nil, domain.ErrInvariant
	}
	id, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	addrs, err := s.Deps.ListAddresses(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]*AddressMsg, 0, len(addrs))
	for _, a := range addrs {
		out = append(out, &AddressMsg{
			Id: a.ID.String(), ProfileId: a.ProfileID.String(), Label: string(a.Label),
			Line1: a.Line1, Lat: a.Lat, Lng: a.Lng, IsDefault: a.IsDefault,
		})
	}
	return &ListAddressesResponse{Addresses: out}, nil
}

// CheckConsent reports whether a consent channel is granted.
func (s *Server) CheckConsent(ctx context.Context, req *CheckConsentRequest) (*CheckConsentResponse, error) {
	if s.Deps == nil {
		return nil, domain.ErrInvariant
	}
	id, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	ok, err := s.Deps.CheckConsent(ctx, id, domain.ConsentChannel(req.Channel))
	if err != nil {
		return nil, err
	}
	return &CheckConsentResponse{Granted: ok}, nil
}

// GetPersonalization returns personalization for a profile.
func (s *Server) GetPersonalization(ctx context.Context, req *GetPersonalizationRequest) (*GetPersonalizationResponse, error) {
	if s.Deps == nil {
		return nil, domain.ErrInvariant
	}
	id, err := uuid.Parse(req.ProfileID)
	if err != nil {
		return nil, domain.ErrInvalidArgument
	}
	p, err := s.Deps.GetPersonalization(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetPersonalizationResponse{Personalization: &PersonalizationMsg{
		ProfileId: p.ProfileID.String(),
	}}, nil
}

func toProfileMsg(p domain.CustomerProfile) *ProfileMsg {
	return &ProfileMsg{
		Id: p.ID.String(), PrincipalId: p.PrincipalID.String(), TenantId: p.TenantID.String(),
		DisplayName: p.DisplayName, FullName: p.FullName, Language: p.Language,
		CountryCode: p.CountryCode, City: p.City, Status: string(p.Status), AvatarUrl: p.AvatarURL,
	}
}
