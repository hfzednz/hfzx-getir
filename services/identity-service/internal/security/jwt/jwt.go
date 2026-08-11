// Package jwt provides RS256 key management, access-token issue/validate, and JWKS export.
package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrNoKey           = errors.New("jwt: no signing key loaded")
	ErrInvalidToken    = errors.New("jwt: invalid token")
	ErrUnexpectedAlg   = errors.New("jwt: unexpected signing method")
	ErrMissingClaims   = errors.New("jwt: missing required claims")
)

// AccessClaims are NEXORA access-token claims.
type AccessClaims struct {
	Subject  string   `json:"sub"`
	Session  string   `json:"sid"`
	Tenant   string   `json:"tid"`
	Roles    []string `json:"roles"`
	AMR      []string `json:"amr"` // authentication methods references
	ACR      string   `json:"acr"` // authentication context class reference
	DeviceID string   `json:"device_id"`
	Issuer   string   `json:"iss"`
	Audience string   `json:"aud"`
	Expires  time.Time
	IssuedAt time.Time
	JTI      string
}

// RegisteredClaims maps to jwt.RegisteredClaims for signing.
func (c AccessClaims) toMapClaims() jwt.MapClaims {
	now := c.IssuedAt
	if now.IsZero() {
		now = time.Now().UTC()
	}
	exp := c.Expires
	mc := jwt.MapClaims{
		"sub":       c.Subject,
		"sid":       c.Session,
		"tid":       c.Tenant,
		"roles":     c.Roles,
		"amr":       c.AMR,
		"acr":       c.ACR,
		"device_id": c.DeviceID,
		"iss":       c.Issuer,
		"aud":       c.Audience,
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
	}
	if c.JTI != "" {
		mc["jti"] = c.JTI
	} else {
		mc["jti"] = uuid.NewString()
	}
	return mc
}

// KeyManager holds an RSA key pair for RS256.
type KeyManager struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
}

// NewKeyManager returns an empty KeyManager; call Load or Generate before use.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// Generate creates a new RSA key pair (bits typically 2048 or 4096) and assigns a kid.
func (km *KeyManager) Generate(bits int) error {
	if bits < 2048 {
		bits = 2048
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("jwt: generate key: %w", err)
	}
	km.mu.Lock()
	defer km.mu.Unlock()
	km.privateKey = key
	km.publicKey = &key.PublicKey
	km.kid = uuid.NewString()
	return nil
}

// LoadPEM loads a PKCS#1 or PKCS#8 private key PEM from disk and derives the public key.
func (km *KeyManager) LoadPEM(path string, kid string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("jwt: read key: %w", err)
	}
	return km.LoadPEMBytes(data, kid)
}

// LoadPEMBytes loads a private key from PEM bytes.
func (km *KeyManager) LoadPEMBytes(pemBytes []byte, kid string) error {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return errors.New("jwt: no PEM block found")
	}
	var key *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("jwt: parse PKCS1: %w", err)
		}
		key = k
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("jwt: parse PKCS8: %w", err)
		}
		var ok bool
		key, ok = parsed.(*rsa.PrivateKey)
		if !ok {
			return errors.New("jwt: PKCS8 key is not RSA")
		}
	default:
		return fmt.Errorf("jwt: unsupported PEM type %q", block.Type)
	}
	if kid == "" {
		kid = uuid.NewString()
	}
	km.mu.Lock()
	defer km.mu.Unlock()
	km.privateKey = key
	km.publicKey = &key.PublicKey
	km.kid = kid
	return nil
}

// KID returns the current key id.
func (km *KeyManager) KID() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.kid
}

// PublicKey returns the RSA public key, or nil if unset.
func (km *KeyManager) PublicKey() *rsa.PublicKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.publicKey
}

// PrivateKeyPEM exports the private key as PKCS#8 PEM (for secure storage / rotation tooling).
func (km *KeyManager) PrivateKeyPEM() ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if km.privateKey == nil {
		return nil, ErrNoKey
	}
	b, err := x509.MarshalPKCS8PrivateKey(km.privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: b}), nil
}

// IssueAccessToken signs AccessClaims as an RS256 JWT.
func (km *KeyManager) IssueAccessToken(claims AccessClaims) (string, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if km.privateKey == nil {
		return "", ErrNoKey
	}
	if claims.Subject == "" || claims.Session == "" || claims.Expires.IsZero() {
		return "", ErrMissingClaims
	}
	if claims.IssuedAt.IsZero() {
		claims.IssuedAt = time.Now().UTC()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims.toMapClaims())
	token.Header["kid"] = km.kid
	signed, err := token.SignedString(km.privateKey)
	if err != nil {
		return "", fmt.Errorf("jwt: sign: %w", err)
	}
	return signed, nil
}

// ParseAndValidate parses an RS256 access token and returns AccessClaims.
func (km *KeyManager) ParseAndValidate(tokenString string, expectedIssuer, expectedAudience string) (*AccessClaims, error) {
	km.mu.RLock()
	pub := km.publicKey
	km.mu.RUnlock()
	if pub == nil {
		return nil, ErrNoKey
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
	}
	if expectedIssuer != "" {
		opts = append(opts, jwt.WithIssuer(expectedIssuer))
	}
	if expectedAudience != "" {
		opts = append(opts, jwt.WithAudience(expectedAudience))
	}

	parsed, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, ErrUnexpectedAlg
		}
		return pub, nil
	}, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}
	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return mapToAccessClaims(mc)
}

func mapToAccessClaims(mc jwt.MapClaims) (*AccessClaims, error) {
	c := &AccessClaims{}
	c.Subject, _ = mc["sub"].(string)
	c.Session, _ = mc["sid"].(string)
	c.Tenant, _ = mc["tid"].(string)
	c.ACR, _ = mc["acr"].(string)
	c.DeviceID, _ = mc["device_id"].(string)
	c.Issuer, _ = mc["iss"].(string)
	c.JTI, _ = mc["jti"].(string)

	switch aud := mc["aud"].(type) {
	case string:
		c.Audience = aud
	case []any:
		if len(aud) > 0 {
			if s, ok := aud[0].(string); ok {
				c.Audience = s
			}
		}
	}

	c.Roles = stringSlice(mc["roles"])
	c.AMR = stringSlice(mc["amr"])

	if exp, ok := asInt64(mc["exp"]); ok {
		c.Expires = time.Unix(exp, 0).UTC()
	}
	if iat, ok := asInt64(mc["iat"]); ok {
		c.IssuedAt = time.Unix(iat, 0).UTC()
	}
	if c.Subject == "" || c.Session == "" {
		return nil, ErrMissingClaims
	}
	return c, nil
}

func stringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

func asInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case int64:
		return n, true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

// JWK is a single JSON Web Key (RSA public).
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS is a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// ExportJWKS returns the public key as a JWKS document.
func (km *KeyManager) ExportJWKS() (*JWKS, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if km.publicKey == nil {
		return nil, ErrNoKey
	}
	n := base64.RawURLEncoding.EncodeToString(km.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(km.publicKey.E)).Bytes())
	return &JWKS{
		Keys: []JWK{{
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
			Kid: km.kid,
			N:   n,
			E:   e,
		}},
	}, nil
}

// ExportJWKSJSON returns JWKS as JSON bytes.
func (km *KeyManager) ExportJWKSJSON() ([]byte, error) {
	jwks, err := km.ExportJWKS()
	if err != nil {
		return nil, err
	}
	return json.Marshal(jwks)
}
