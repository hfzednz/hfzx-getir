package domain

import (
	"time"

	"github.com/google/uuid"
)

// DevicePlatform identifies push platform.
type DevicePlatform string

const (
	PlatformFCM  DevicePlatform = "fcm"
	PlatformAPNs DevicePlatform = "apns"
	PlatformWeb  DevicePlatform = "web"
)

// Device holds a push token registration.
type Device struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	PrincipalID uuid.UUID
	Platform    DevicePlatform
	Token       string
	Locale      string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
