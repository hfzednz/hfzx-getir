package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// Sort modes.
const (
	SortRelevance     = "relevance"
	SortPopularity    = "popularity"
	SortNewest        = "newest"
	SortPriceAsc      = "price_asc"
	SortPriceDesc     = "price_desc"
	SortRating        = "rating"
	SortDiscount      = "discount"
	SortDeliverySpeed = "delivery_speed"
	SortAI            = "ai"
)

// Intent types.
const (
	IntentBrowse  = "browse"
	IntentFind    = "find"
	IntentCompare = "compare"
	IntentDeal    = "deal"
	IntentUnknown = "unknown"
)

// Index job modes.
const (
	IndexFull        = "full"
	IndexIncremental = "incremental"
	IndexRealtime    = "realtime"
	IndexBatch       = "batch"
)

// ProductDocument is a denormalized search document (not catalog SoT).
type ProductDocument struct {
	TenantID       uuid.UUID
	ProductID      uuid.UUID
	VariantID      uuid.UUID
	SKU            string
	Title          string
	TitleLocalized map[string]string
	Description    string
	BrandID        uuid.UUID
	BrandName      string
	CategoryIDs    []uuid.UUID
	CategoryPath   []string
	Tags           []string
	Attributes     map[string]string // diet, organic, nutrition keys
	PriceMinor     int64
	CompareAtMinor int64
	DiscountPct    float64
	Currency       string
	Available      bool
	WarehouseIDs   []uuid.UUID
	CityID         uuid.UUID
	RatingAvg      float64
	ReviewCount    int
	Popularity     float64
	FreshnessScore float64
	ProfitScore    float64
	DeliveryETAMin int
	ImageRef       string
	Embedding      []float64
	IndexedAt      time.Time
	Version        int64
}

// SearchFilters constrains results.
type SearchFilters struct {
	CategoryIDs  []uuid.UUID
	BrandIDs     []uuid.UUID
	PriceMin     *int64
	PriceMax     *int64
	DiscountMin  *float64
	AvailableOnly bool
	WarehouseID  *uuid.UUID
	RatingMin    *float64
	ReviewMin    *int
	MaxETAMin    *int
	Attrs        map[string]string // organic=true, diet=vegan
	LocalOnly    bool
}

// SearchQuery is a normalized request.
type SearchQuery struct {
	TenantID       uuid.UUID
	RawQuery       string
	Normalized     string
	Locale         string
	CityID         *uuid.UUID
	UserID         *uuid.UUID
	Device         string
	Filters        SearchFilters
	Sort           string
	Page           int
	PageSize       int
	Hybrid         bool
	IncludeFacets  bool
	Personalize    bool
}

// RankedHit is a scored document id.
type RankedHit struct {
	ProductID      uuid.UUID
	VariantID      uuid.UUID
	Score          float64
	LexicalScore   float64
	VectorScore    float64
	BehaviorScore  float64
	MerchBoost     float64
	Sponsored      bool
	Pinned         bool
	Reasons        []string
}

// SearchResult is the API search response model.
type SearchResult struct {
	QueryID        uuid.UUID
	RewrittenQuery string
	DidYouMean     string
	Intent         string
	Hits           []RankedHit
	Total          int
	Facets         map[string][]FacetBucket
	TookMs         int64
	ZeroResult     bool
}

// FacetBucket is a facet count.
type FacetBucket struct {
	Key   string
	Count int
}

// SynonymRule expands query terms.
type SynonymRule struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Locale    string
	Term      string
	Synonyms  []string
	Active    bool
	UpdatedAt time.Time
}

// BoostRule merchandising pin/boost/demote.
type BoostRule struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Kind       string // pin|boost|demote|sponsor
	ProductIDs []uuid.UUID
	CategoryID *uuid.UUID
	BrandID    *uuid.UUID
	Weight     float64
	Priority   int
	Active     bool
	StartsAt   *time.Time
	EndsAt     *time.Time
	UpdatedAt  time.Time
}

// IndexJob tracks indexing work.
type IndexJob struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Mode      string
	Status    string // pending|running|done|failed
	DocsTotal int
	DocsDone  int
	Error     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TrendEntry is a trending entity.
type TrendEntry struct {
	TenantID   uuid.UUID
	Kind       string // search|product|category
	Key        string
	EntityID   *uuid.UUID
	Score      float64
	Region     string
	UpdatedAt  time.Time
}

// SuggestCandidate autocomplete row.
type SuggestCandidate struct {
	Text       string
	ProductID  *uuid.UUID
	CategoryID *uuid.UUID
	Weight     float64
}

// NormalizeQuery lowercases, trims, collapses spaces.
func NormalizeQuery(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	var b strings.Builder
	prevSpace := false
	for _, r := range q {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

// DetectIntent crude intent classifier.
func DetectIntent(q string) string {
	q = strings.ToLower(q)
	switch {
	case strings.Contains(q, "vs ") || strings.Contains(q, "compare"):
		return IntentCompare
	case strings.Contains(q, "deal") || strings.Contains(q, "discount") || strings.Contains(q, "indirim"):
		return IntentDeal
	case q == "" || q == "*":
		return IntentBrowse
	default:
		return IntentFind
	}
}

// ValidSort reports supported sort.
func ValidSort(s string) bool {
	switch s {
	case "", SortRelevance, SortPopularity, SortNewest, SortPriceAsc, SortPriceDesc,
		SortRating, SortDiscount, SortDeliverySpeed, SortAI:
		return true
	default:
		return false
	}
}

// Levenshtein distance for did-you-mean / fuzzy.
func Levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = minInt(del, minInt(ins, sub))
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Tokenize splits on whitespace.
func Tokenize(q string) []string {
	parts := strings.Fields(NormalizeQuery(q))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if IsStopWord(p) {
			continue
		}
		out = append(out, Stem(p))
	}
	return out
}

// IsStopWord basic TR/EN stop list.
func IsStopWord(w string) bool {
	switch w {
	case "a", "an", "the", "and", "or", "of", "in", "on", "for",
		"bir", "ve", "ile", "icin", "için", "bu", "şu", "o":
		return true
	default:
		return false
	}
}

// Stem naive suffix trim (production: snowball).
func Stem(w string) string {
	for _, suf := range []string{"ing", "ed", "ler", "lar", "ları", "leri"} {
		if strings.HasSuffix(w, suf) && len(w) > len(suf)+2 {
			return strings.TrimSuffix(w, suf)
		}
	}
	return w
}

// CosineSimilarity for embedding vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += a[i] * b[i]
		na += a[i] * a[i]
		nb += b[i] * b[i]
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (sqrt(na) * sqrt(nb))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 12; i++ {
		z = 0.5 * (z + x/z)
	}
	return z
}

// ReciprocalRankFusion merges ranked lists.
func ReciprocalRankFusion(lists [][]uuid.UUID, k float64) map[uuid.UUID]float64 {
	if k <= 0 {
		k = 60
	}
	scores := map[uuid.UUID]float64{}
	for _, list := range lists {
		for rank, id := range list {
			scores[id] += 1.0 / (k + float64(rank+1))
		}
	}
	return scores
}

// ComputeRankScore blends component scores.
func ComputeRankScore(lexical, vector, popularity, ctr, conversion, inventory, profit, freshness, behavior, merch float64) float64 {
	return lexical*0.28 + vector*0.22 + popularity*0.1 + ctr*0.08 + conversion*0.08 +
		inventory*0.07 + profit*0.05 + freshness*0.04 + behavior*0.05 + merch*0.03
}
