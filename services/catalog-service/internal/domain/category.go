package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// CategoryKind classifies taxonomy nodes.
type CategoryKind string

const (
	CategoryKindDepartment CategoryKind = "department"
	CategoryKindStandard   CategoryKind = "standard"
	CategoryKindSmart      CategoryKind = "smart"
	CategoryKindCollection CategoryKind = "collection"
	CategoryKindCampaign   CategoryKind = "campaign"
	CategoryKindFeatured   CategoryKind = "featured"
)

func (k CategoryKind) Valid() bool {
	switch k {
	case CategoryKindDepartment, CategoryKindStandard, CategoryKindSmart,
		CategoryKindCollection, CategoryKindCampaign, CategoryKindFeatured:
		return true
	default:
		return false
	}
}

const (
	maxCategoryNameLen = 200
	maxCategoryDepth   = 32
)

// Category is a node in the tenant product taxonomy tree.
type Category struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Kind        CategoryKind
	Path        string
	Depth       int
	SortOrder   int
	Description string
	ImageURL    string
	Metadata    map[string]any
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Validate checks structural invariants (not tree cycle — use WouldCreateCategoryCycle).
func (c Category) Validate() error {
	if c.ID == uuid.Nil {
		return fmt.Errorf("%w: category id required", ErrInvalidArgument)
	}
	if c.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: category name required", ErrInvalidArgument)
	}
	if utf8.RuneCountInString(c.Name) > maxCategoryNameLen {
		return fmt.Errorf("%w: category name too long", ErrInvalidArgument)
	}
	if err := ValidateSlug(c.Slug); err != nil {
		return err
	}
	if !c.Kind.Valid() {
		return fmt.Errorf("%w: invalid category kind %q", ErrInvalidArgument, c.Kind)
	}
	if c.Depth < 0 || c.Depth > maxCategoryDepth {
		return fmt.Errorf("%w: category depth out of range", ErrInvalidArgument)
	}
	if c.ParentID != nil {
		if *c.ParentID == uuid.Nil {
			return fmt.Errorf("%w: parent_id must not be nil UUID", ErrInvalidArgument)
		}
		if *c.ParentID == c.ID {
			return fmt.Errorf("%w: category cannot be its own parent", ErrCategoryCycle)
		}
	}
	return nil
}

// ProductCategory links a product into a category.
type ProductCategory struct {
	ProductID  uuid.UUID
	CategoryID uuid.UUID
	IsPrimary  bool
	SortOrder  int
	AssignedAt time.Time
}

// Validate checks structural invariants.
func (pc ProductCategory) Validate() error {
	if pc.ProductID == uuid.Nil {
		return fmt.Errorf("%w: product_id required", ErrInvalidArgument)
	}
	if pc.CategoryID == uuid.Nil {
		return fmt.Errorf("%w: category_id required", ErrInvalidArgument)
	}
	return nil
}

// BuildCategoryPath builds a materialized path from parent path + self id.
func BuildCategoryPath(parentPath string, id uuid.UUID) string {
	seg := id.String()
	parentPath = strings.TrimSpace(parentPath)
	if parentPath == "" || parentPath == "/" {
		return "/" + seg
	}
	return strings.TrimRight(parentPath, "/") + "/" + seg
}

// WouldCreateCategoryCycle reports whether setting childID's parent to newParentID
// would introduce a cycle, given ancestorIDsOfParent (root→…→newParent inclusive).
// Pass the full ancestor chain of the prospective parent (including itself).
func WouldCreateCategoryCycle(childID, newParentID uuid.UUID, ancestorIDsOfParent []uuid.UUID) bool {
	if childID == uuid.Nil || newParentID == uuid.Nil {
		return false
	}
	if childID == newParentID {
		return true
	}
	for _, a := range ancestorIDsOfParent {
		if a == childID {
			return true
		}
	}
	return false
}

// ParseCategoryPathIDs extracts UUIDs from a materialized path string.
func ParseCategoryPathIDs(path string) ([]uuid.UUID, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil, nil
	}
	parts := strings.Split(path, "/")
	out := make([]uuid.UUID, 0, len(parts))
	for _, p := range parts {
		id, err := uuid.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid path segment %q", ErrInvalidArgument, p)
		}
		out = append(out, id)
	}
	return out, nil
}
