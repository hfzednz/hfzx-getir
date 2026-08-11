package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type ModuleKind string

const (
	KindMiniApp ModuleKind = "mini_app"
	KindPlugin  ModuleKind = "plugin"
	KindWidget  ModuleKind = "widget"
	KindExt     ModuleKind = "extension"
)

type ModuleStatus string

const (
	ModuleDraft     ModuleStatus = "draft"
	ModulePublished ModuleStatus = "published"
	ModuleSuspended ModuleStatus = "suspended"
	ModuleRetired   ModuleStatus = "retired"
)

// Module is the registry entry for mini-apps, plugins, widgets.
type Module struct {
	ID            uuid.UUID    `json:"id"`
	TenantID      uuid.UUID    `json:"tenantId"`
	Key           string       `json:"key"`
	Name          string       `json:"name"`
	Kind          ModuleKind   `json:"kind"`
	Category      string       `json:"category"`
	Description   string       `json:"description"`
	PublisherID   string       `json:"publisherId"` // opaque partner/dev ref
	Status        ModuleStatus `json:"status"`
	LatestVersion string       `json:"latestVersion"`
	IconURI       string       `json:"iconUri"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

type ModuleManifest struct {
	ID               uuid.UUID         `json:"id"`
	TenantID         uuid.UUID         `json:"tenantId"`
	ModuleID         uuid.UUID         `json:"moduleId"`
	Version          string            `json:"version"`
	EntryPoint       string            `json:"entryPoint"` // flutter:// | wasm:// | mf://
	MinShellVersion  string            `json:"minShellVersion"`
	Permissions      []string          `json:"permissions"`
	Hooks            []string          `json:"hooks"` // navigation, payment, search, ...
	Dependencies     map[string]string `json:"dependencies"`
	Signature        string            `json:"signature"`
	Checksum         string            `json:"checksum"`
	BundleURI        string            `json:"bundleUri"`
	Compatible       bool              `json:"compatible"`
	CreatedAt        time.Time         `json:"createdAt"`
}

type InstallStatus string

const (
	InstallPending   InstallStatus = "pending"
	InstallActive    InstallStatus = "active"
	InstallUpdating  InstallStatus = "updating"
	InstallRemoved   InstallStatus = "removed"
	InstallRolledBack InstallStatus = "rolled_back"
)

type Install struct {
	ID             uuid.UUID     `json:"id"`
	TenantID       uuid.UUID     `json:"tenantId"`
	SubjectID      string        `json:"subjectId"` // user or device opaque
	ModuleID       uuid.UUID     `json:"moduleId"`
	Version        string        `json:"version"`
	Status         InstallStatus `json:"status"`
	PreviousVersion string       `json:"previousVersion,omitempty"`
	InstalledAt    time.Time     `json:"installedAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type PermissionGrant struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	SubjectID  string    `json:"subjectId"`
	ModuleID   uuid.UUID `json:"moduleId"`
	Permission string    `json:"permission"`
	Granted    bool      `json:"granted"`
	GrantedAt  time.Time `json:"grantedAt"`
}

type StoreListing struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	ModuleID    uuid.UUID `json:"moduleId"`
	Featured    bool      `json:"featured"`
	PriceMinor  int64     `json:"priceMinor"`
	Currency    string    `json:"currency"`
	Subscription bool     `json:"subscription"`
	RatingAvg   float64   `json:"ratingAvg"`
	RatingCount int       `json:"ratingCount"`
	Installs    int64     `json:"installs"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type StoreRating struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	ModuleID  uuid.UUID `json:"moduleId"`
	SubjectID string    `json:"subjectId"`
	Score     int       `json:"score"` // 1..5
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type WidgetPlacement struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	SubjectID string    `json:"subjectId"`
	ModuleID  uuid.UUID `json:"moduleId"`
	Slot      string    `json:"slot"` // home|dashboard|notification|live|ai|recommendation
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
}

type MonetizationRule struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenantId"`
	ModuleID         uuid.UUID `json:"moduleId"`
	CommissionBps    int       `json:"commissionBps"`
	PartnerShareBps  int       `json:"partnerShareBps"`
	Active           bool      `json:"active"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type LaunchEvent struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	SubjectID string    `json:"subjectId"`
	ModuleID  uuid.UUID `json:"moduleId"`
	Kind      ModuleKind `json:"kind"`
	CreatedAt time.Time `json:"createdAt"`
}

// ShellResolve is returned to the Flutter shell for dynamic loading.
type ShellResolve struct {
	Modules   []ResolvedModule `json:"modules"`
	Widgets   []WidgetPlacement `json:"widgets"`
	ShellHint string           `json:"shellHint"`
}

type ResolvedModule struct {
	Key         string   `json:"key"`
	Kind        ModuleKind `json:"kind"`
	Version     string   `json:"version"`
	EntryPoint  string   `json:"entryPoint"`
	Permissions []string `json:"permissions"`
	Hooks       []string `json:"hooks"`
	BundleURI   string   `json:"bundleUri"`
}

const (
	PermPayments     = "payments"
	PermNotifications = "notifications"
	PermProfile      = "profile"
	PermOrders       = "orders"
	PermSearch       = "search"
	PermAnalytics    = "analytics"
	PermAI           = "ai"
	PermNavigation   = "navigation"
)

func ValidateModule(m Module) error {
	if m.TenantID == uuid.Nil || strings.TrimSpace(m.Key) == "" || m.Name == "" || m.Kind == "" {
		return ErrInvalidArgument
	}
	return nil
}

func ValidateManifest(man ModuleManifest) error {
	if man.TenantID == uuid.Nil || man.ModuleID == uuid.Nil || man.Version == "" || man.EntryPoint == "" {
		return ErrInvalidArgument
	}
	if man.Signature == "" || man.Checksum == "" {
		return ErrNotSigned
	}
	return nil
}

func SemverCompatible(shell, minRequired string) bool {
	// simplified major.minor compare for shell readiness
	if minRequired == "" {
		return true
	}
	sp := strings.Split(shell, ".")
	mp := strings.Split(minRequired, ".")
	for i := 0; i < 2; i++ {
		sv, mv := 0, 0
		if i < len(sp) {
			sv = atoi(sp[i])
		}
		if i < len(mp) {
			mv = atoi(mp[i])
		}
		if sv > mv {
			return true
		}
		if sv < mv {
			return false
		}
	}
	return true
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func DefaultMiniAppCatalog() []Module {
	defs := []struct{ key, name, cat string }{
		{"qc", "Quick Commerce", "commerce"},
		{"food", "Food Delivery", "food"},
		{"pharmacy", "Pharmacy", "health"},
		{"flowers", "Flowers", "gifts"},
		{"pets", "Pet Shop", "pets"},
		{"electronics", "Electronics", "retail"},
		{"gifts", "Gift Shop", "gifts"},
		{"tickets", "Ticketing", "lifestyle"},
		{"travel", "Travel", "travel"},
		{"ride", "Ride Hailing", "mobility"},
		{"bills", "Bill Payments", "finance"},
		{"insurance", "Insurance", "finance"},
		{"banking", "Banking", "finance"},
		{"health", "Health", "health"},
		{"gov", "Government", "public"},
		{"marketplace", "Marketplace", "commerce"},
		{"corporate", "Corporate Portal", "b2b"},
	}
	out := make([]Module, 0, len(defs))
	for _, d := range defs {
		out = append(out, Module{Key: d.key, Name: d.name, Kind: KindMiniApp, Category: d.cat, Status: ModulePublished, LatestVersion: "1.0.0"})
	}
	return out
}

func AllowedPermissions() map[string]bool {
	return map[string]bool{
		PermPayments: true, PermNotifications: true, PermProfile: true, PermOrders: true,
		PermSearch: true, PermAnalytics: true, PermAI: true, PermNavigation: true,
	}
}

func ValidatePermissions(perms []string) error {
	allow := AllowedPermissions()
	for _, p := range perms {
		if !allow[p] {
			return ErrSandboxViolation
		}
	}
	return nil
}

func UpdateRatingAvg(avg float64, count int, score int) (float64, int) {
	n := count + 1
	return (avg*float64(count) + float64(score)) / float64(n), n
}
