// Package webauthn defines WebAuthn/FIDO2 registration and login interfaces for NEXORA.
//
// Production wiring: replace StubService with an adapter around
// github.com/go-webauthn/webauthn (or equivalent). Struct shapes below mirror
// the WebAuthn Level 2 ceremony data so the HTTP/app layer stays library-agnostic.
package webauthn

import (
	"context"
	"encoding/base64"
	"errors"
	"time"
)

var (
	ErrNotImplemented = errors.New("webauthn: not implemented — wire go-webauthn in production")
	ErrInvalidSession = errors.New("webauthn: invalid or expired ceremony session")
	ErrVerification   = errors.New("webauthn: credential verification failed")
)

// User identifies a WebAuthn user entity.
type User struct {
	ID          []byte
	Name        string // typically email or username
	DisplayName string
}

// Credential is a stored public-key credential.
type Credential struct {
	ID              []byte
	PublicKey       []byte
	AttestationType string
	Transport       []string // usb, nfc, ble, internal, hybrid
	AAGUID          []byte
	SignCount       uint32
	BackupEligible  bool
	BackupState     bool
	CreatedAt       time.Time
}

// PublicKeyCredentialCreationOptions is the RP → client registration challenge payload.
type PublicKeyCredentialCreationOptions struct {
	Challenge              []byte                 `json:"challenge"`
	RelyingParty           RelyingPartyEntity     `json:"rp"`
	User                   UserEntity             `json:"user"`
	PubKeyCredParams       []PubKeyCredParam      `json:"pubKeyCredParams"`
	Timeout                int                    `json:"timeout,omitempty"`
	ExcludeCredentials     []CredentialDescriptor `json:"excludeCredentials,omitempty"`
	AuthenticatorSelection *AuthenticatorSelection `json:"authenticatorSelection,omitempty"`
	Attestation            string                 `json:"attestation,omitempty"`
	SessionID              string                 `json:"-"` // server-side ceremony id
}

// PublicKeyCredentialRequestOptions is the RP → client authentication challenge payload.
type PublicKeyCredentialRequestOptions struct {
	Challenge        []byte                 `json:"challenge"`
	Timeout          int                    `json:"timeout,omitempty"`
	RelyingPartyID   string                 `json:"rpId,omitempty"`
	AllowCredentials []CredentialDescriptor `json:"allowCredentials,omitempty"`
	UserVerification string                 `json:"userVerification,omitempty"`
	SessionID        string                 `json:"-"`
}

// RelyingPartyEntity describes the RP.
type RelyingPartyEntity struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// UserEntity is the WebAuthn user account descriptor.
type UserEntity struct {
	ID          string `json:"id"` // base64url of User.ID
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// PubKeyCredParam is an allowed COSE algorithm.
type PubKeyCredParam struct {
	Type string `json:"type"` // "public-key"
	Alg  int    `json:"alg"`  // e.g. -7 (ES256), -257 (RS256)
}

// CredentialDescriptor identifies an existing credential.
type CredentialDescriptor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"` // base64url
	Transports []string `json:"transports,omitempty"`
}

// AuthenticatorSelection constrains authenticator choice.
type AuthenticatorSelection struct {
	AuthenticatorAttachment string `json:"authenticatorAttachment,omitempty"`
	RequireResidentKey      bool   `json:"requireResidentKey,omitempty"`
	ResidentKey             string `json:"residentKey,omitempty"`
	UserVerification        string `json:"userVerification,omitempty"`
}

// AuthenticatorAttestationResponse is the client → RP registration result.
type AuthenticatorAttestationResponse struct {
	ID       string `json:"id"`
	RawID    []byte `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    []byte `json:"clientDataJSON"`
		AttestationObject []byte `json:"attestationObject"`
		Transports        []string `json:"transports,omitempty"`
	} `json:"response"`
	SessionID string `json:"sessionId"`
}

// AuthenticatorAssertionResponse is the client → RP login result.
type AuthenticatorAssertionResponse struct {
	ID       string `json:"id"`
	RawID    []byte `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    []byte `json:"clientDataJSON"`
		AuthenticatorData []byte `json:"authenticatorData"`
		Signature         []byte `json:"signature"`
		UserHandle        []byte `json:"userHandle,omitempty"`
	} `json:"response"`
	SessionID string `json:"sessionId"`
}

// CeremonySession is ephemeral state between Begin* and Finish*.
type CeremonySession struct {
	ID        string
	UserID    []byte
	Challenge []byte
	Type      string // "registration" | "login"
	ExpiresAt time.Time
}

// Service is the WebAuthn ceremony port.
type Service interface {
	BeginRegistration(ctx context.Context, user User, exclude []Credential) (*PublicKeyCredentialCreationOptions, *CeremonySession, error)
	FinishRegistration(ctx context.Context, session *CeremonySession, resp *AuthenticatorAttestationResponse) (*Credential, error)
	BeginLogin(ctx context.Context, user User, allowed []Credential) (*PublicKeyCredentialRequestOptions, *CeremonySession, error)
	FinishLogin(ctx context.Context, session *CeremonySession, resp *AuthenticatorAssertionResponse) (*Credential, error)
}

// Config holds RP settings used when wiring a real library.
type Config struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
	Timeout       time.Duration
}

// StubService implements Service by returning ErrNotImplemented.
// Use it in tests / early boot until go-webauthn is wired.
type StubService struct {
	Config Config
}

// BeginRegistration implements Service.
func (s *StubService) BeginRegistration(_ context.Context, user User, _ []Credential) (*PublicKeyCredentialCreationOptions, *CeremonySession, error) {
	if user.Name == "" {
		return nil, nil, errors.New("webauthn: user name required")
	}
	return nil, nil, ErrNotImplemented
}

// FinishRegistration implements Service.
func (s *StubService) FinishRegistration(_ context.Context, _ *CeremonySession, _ *AuthenticatorAttestationResponse) (*Credential, error) {
	return nil, ErrNotImplemented
}

// BeginLogin implements Service.
func (s *StubService) BeginLogin(_ context.Context, user User, _ []Credential) (*PublicKeyCredentialRequestOptions, *CeremonySession, error) {
	if user.Name == "" {
		return nil, nil, errors.New("webauthn: user name required")
	}
	return nil, nil, ErrNotImplemented
}

// FinishLogin implements Service.
func (s *StubService) FinishLogin(_ context.Context, _ *CeremonySession, _ *AuthenticatorAssertionResponse) (*Credential, error) {
	return nil, ErrNotImplemented
}

// EncodeCredentialID returns base64url (no padding) of a credential id.
func EncodeCredentialID(id []byte) string {
	return base64.RawURLEncoding.EncodeToString(id)
}

// DecodeCredentialID parses a base64url credential id.
func DecodeCredentialID(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// Ensure StubService satisfies Service at compile time.
var _ Service = (*StubService)(nil)
