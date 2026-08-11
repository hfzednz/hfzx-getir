package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
)

// OIDCIdP performs authorization-code exchange against a standard OIDC token + userinfo endpoint.
type OIDCIdP struct {
	provider     ports.SocialProvider
	clientID     string
	clientSecret string
	tokenURL     string
	userInfoURL  string
	httpClient   *http.Client
}

func (p *OIDCIdP) Provider() ports.SocialProvider { return p.provider }

func (p *OIDCIdP) Exchange(ctx context.Context, code, redirectURI string) (ports.SocialProfile, error) {
	if p.clientID == "" || p.clientSecret == "" {
		return ports.SocialProfile{}, fmt.Errorf("%w: social provider %s not configured", domain.ErrUnauthorized, p.provider)
	}
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ports.SocialProfile{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return ports.SocialProfile{}, fmt.Errorf("social: token exchange: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ports.SocialProfile{}, fmt.Errorf("%w: token endpoint %d: %s", domain.ErrUnauthorized, resp.StatusCode, string(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return ports.SocialProfile{}, err
	}
	if tok.AccessToken == "" {
		return ports.SocialProfile{}, fmt.Errorf("%w: missing access_token", domain.ErrUnauthorized)
	}

	ureq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return ports.SocialProfile{}, err
	}
	ureq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	ureq.Header.Set("Accept", "application/json")
	uresp, err := p.httpClient.Do(ureq)
	if err != nil {
		return ports.SocialProfile{}, fmt.Errorf("social: userinfo: %w", err)
	}
	defer uresp.Body.Close()
	ubody, _ := io.ReadAll(io.LimitReader(uresp.Body, 1<<20))
	if uresp.StatusCode < 200 || uresp.StatusCode >= 300 {
		return ports.SocialProfile{}, fmt.Errorf("%w: userinfo %d: %s", domain.ErrUnauthorized, uresp.StatusCode, string(ubody))
	}
	var u struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"`
		Name          string `json:"name"`
		Preferred     string `json:"preferred_username"`
	}
	if err := json.Unmarshal(ubody, &u); err != nil {
		return ports.SocialProfile{}, err
	}
	if u.Sub == "" {
		return ports.SocialProfile{}, fmt.Errorf("%w: missing sub", domain.ErrUnauthorized)
	}
	verified := false
	switch v := u.EmailVerified.(type) {
	case bool:
		verified = v
	case string:
		verified = strings.EqualFold(v, "true")
	}
	name := u.Name
	if name == "" {
		name = u.Preferred
	}
	if name == "" {
		name = u.Email
	}
	return ports.SocialProfile{
		Provider: p.provider, Subject: u.Sub, Email: u.Email,
		EmailVerified: verified, DisplayName: name,
	}, nil
}

// LoadProvidersFromEnv builds production IdPs for any providers with credentials configured.
func LoadProvidersFromEnv() map[ports.SocialProvider]ports.SocialIdP {
	client := &http.Client{Timeout: 15 * time.Second}
	out := map[ports.SocialProvider]ports.SocialIdP{}
	add := func(provider ports.SocialProvider, idEnv, secretEnv, tokenURL, userInfoURL string) {
		id := os.Getenv(idEnv)
		secret := os.Getenv(secretEnv)
		if id == "" || secret == "" {
			return
		}
		out[provider] = &OIDCIdP{
			provider: provider, clientID: id, clientSecret: secret,
			tokenURL: tokenURL, userInfoURL: userInfoURL, httpClient: client,
		}
	}
	add(ports.SocialGoogle, "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"https://oauth2.googleapis.com/token", "https://openidconnect.googleapis.com/v1/userinfo")
	add(ports.SocialMicrosoft, "MICROSOFT_CLIENT_ID", "MICROSOFT_CLIENT_SECRET",
		"https://login.microsoftonline.com/common/oauth2/v2.0/token", "https://graph.microsoft.com/oidc/userinfo")
	add(ports.SocialGitHub, "GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		"https://github.com/login/oauth/access_token", "https://api.github.com/user")
	add(ports.SocialFacebook, "FACEBOOK_CLIENT_ID", "FACEBOOK_CLIENT_SECRET",
		"https://graph.facebook.com/v19.0/oauth/access_token", "https://graph.facebook.com/me?fields=id,name,email")
	add(ports.SocialApple, "APPLE_CLIENT_ID", "APPLE_CLIENT_SECRET",
		"https://appleid.apple.com/auth/token", "https://appleid.apple.com/auth/userinfo")
	return out
}

var _ ports.SocialIdP = (*OIDCIdP)(nil)
