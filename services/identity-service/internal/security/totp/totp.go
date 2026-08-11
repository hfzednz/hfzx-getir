// Package totp implements RFC 6238 TOTP (HMAC-SHA1, 30s step, 6 digits by default).
package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultDigits    = 6
	DefaultPeriod    = 30
	DefaultSecretLen = 20 // 160-bit
	DefaultSkew      = 1  // ±1 time step
)

var (
	ErrInvalidSecret = errors.New("totp: invalid secret")
	ErrInvalidCode   = errors.New("totp: invalid code")
)

// Options configures TOTP generation and validation.
type Options struct {
	Digits int           // default 6
	Period int64         // seconds, default 30
	Skew   int           // acceptable window steps, default 1
	Issuer string        // for otpauth URI
	Now    func() time.Time
}

func (o Options) normalize() Options {
	if o.Digits <= 0 {
		o.Digits = DefaultDigits
	}
	if o.Period <= 0 {
		o.Period = DefaultPeriod
	}
	if o.Skew < 0 {
		o.Skew = DefaultSkew
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	return o
}

// GenerateSecret returns a new base32-encoded (no padding) shared secret.
func GenerateSecret() (string, error) {
	return GenerateSecretN(DefaultSecretLen)
}

// GenerateSecretN returns a base32 secret from n random bytes.
func GenerateSecretN(n int) (string, error) {
	if n < 10 {
		n = DefaultSecretLen
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("totp: generate secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf), nil
}

// OTPAuthURI builds an otpauth://totp/... provisioning URI.
func OTPAuthURI(secret, accountName string, opts Options) (string, error) {
	opts = opts.normalize()
	if secret == "" || accountName == "" {
		return "", ErrInvalidSecret
	}
	label := accountName
	if opts.Issuer != "" {
		label = opts.Issuer + ":" + accountName
	}
	v := url.Values{}
	v.Set("secret", secret)
	v.Set("algorithm", "SHA1")
	v.Set("digits", fmt.Sprintf("%d", opts.Digits))
	v.Set("period", fmt.Sprintf("%d", opts.Period))
	if opts.Issuer != "" {
		v.Set("issuer", opts.Issuer)
	}
	return "otpauth://totp/" + url.PathEscape(label) + "?" + v.Encode(), nil
}

// GenerateCode computes the TOTP code for secret at the given time.
func GenerateCode(secret string, t time.Time, opts Options) (string, error) {
	opts = opts.normalize()
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix() / opts.Period)
	return hotp(key, counter, opts.Digits), nil
}

// ValidateCode verifies a user-supplied TOTP code within the skew window.
func ValidateCode(secret, code string, opts Options) (bool, error) {
	opts = opts.normalize()
	if code == "" {
		return false, ErrInvalidCode
	}
	key, err := decodeSecret(secret)
	if err != nil {
		return false, err
	}
	now := opts.Now().Unix()
	counter := now / opts.Period
	for i := -opts.Skew; i <= opts.Skew; i++ {
		c := int64(counter) + int64(i)
		if c < 0 {
			continue
		}
		expected := hotp(key, uint64(c), opts.Digits)
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true, nil
		}
	}
	return false, nil
}

func decodeSecret(secret string) ([]byte, error) {
	s := strings.ToUpper(strings.TrimSpace(secret))
	s = strings.ReplaceAll(s, " ", "")
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		// try with padding
		key, err = base32.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, ErrInvalidSecret
		}
	}
	if len(key) == 0 {
		return nil, ErrInvalidSecret
	}
	return key, nil
}

func hotp(key []byte, counter uint64, digits int) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	mod := uint32(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	code := truncated % mod
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, code)
}
