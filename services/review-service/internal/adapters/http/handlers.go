package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app"
	"github.com/nexora/review-service/internal/domain"
	"github.com/nexora/review-service/internal/ratelimit"
)

// Handler serves Review REST endpoints.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/reviews"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"", tenant(h.createReview))
	mux.HandleFunc("GET "+base+"/{id}", tenant(h.getReview))
	mux.HandleFunc("PATCH "+base+"/{id}", tenant(h.updateReview))
	mux.HandleFunc("DELETE "+base+"/{id}", tenant(h.deleteReview))
	mux.HandleFunc("GET "+base+"/targets/{targetType}/{targetId}", tenant(h.listByTarget))
	mux.HandleFunc("GET "+base+"/authors/{authorId}", tenant(h.listByAuthor))
	mux.HandleFunc("GET "+base+"/{id}/revisions", tenant(h.listRevisions))

	mux.HandleFunc("POST "+base+"/ratings", tenant(h.submitRating))
	mux.HandleFunc("GET "+base+"/ratings/{targetType}/{targetId}", tenant(h.getAggregate))

	mux.HandleFunc("POST "+base+"/{id}/votes", tenant(h.vote))
	mux.HandleFunc("POST "+base+"/{id}/comments", tenant(h.comment))
	mux.HandleFunc("GET "+base+"/{id}/comments", tenant(h.listComments))
	mux.HandleFunc("POST "+base+"/{id}/reports", tenant(h.report))
	mux.HandleFunc("POST "+base+"/{id}/pin", tenant(h.pin))

	mux.HandleFunc("GET "+base+"/moderation/queue", tenant(h.moderationQueue))
	mux.HandleFunc("POST "+base+"/moderation/{id}/decide", tenant(h.decideModeration))

	mux.HandleFunc("GET "+base+"/trust/{reviewerId}", tenant(h.getTrust))
	mux.HandleFunc("GET "+base+"/reputation/{targetType}/{targetId}", tenant(h.getReputation))
	mux.HandleFunc("GET "+base+"/ai/summarize/{targetType}/{targetId}", tenant(h.summarize))
	mux.HandleFunc("GET "+base+"/search", tenant(h.search))

	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.adminStats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.publishOutbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ready"})
}

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func parsePathID(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(key))
}

func reviewDTO(rev domain.Review) map[string]any {
	dto := map[string]any{
		"id": rev.ID.String(), "tenantId": rev.TenantID.String(),
		"authorId": rev.AuthorID.String(), "targetType": rev.TargetType,
		"targetId": rev.TargetID.String(), "locale": rev.Locale,
		"title": rev.Title, "body": rev.Body, "anonymous": rev.Anonymous,
		"verifiedPurchase": rev.VerifiedPurchase, "verifiedDelivery": rev.VerifiedDelivery,
		"status": rev.Status, "sentiment": rev.Sentiment, "topics": rev.Topics,
		"tags": rev.Tags, "helpfulCount": rev.HelpfulCount, "notHelpfulCount": rev.NotHelpfulCount,
		"reportCount": rev.ReportCount, "pinned": rev.Pinned, "revision": rev.Revision,
		"createdAt": rev.CreatedAt, "updatedAt": rev.UpdatedAt,
	}
	if rev.OrderID != nil {
		dto["orderId"] = rev.OrderID.String()
	}
	if rev.PublishedAt != nil {
		dto["publishedAt"] = rev.PublishedAt
	}
	return dto
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		AuthorID       uuid.UUID          `json:"authorId"`
		TargetType     string             `json:"targetType"`
		TargetID       uuid.UUID          `json:"targetId"`
		OrderID        *uuid.UUID         `json:"orderId"`
		Locale         string             `json:"locale"`
		Title          string             `json:"title"`
		Body           string             `json:"body"`
		Anonymous      bool               `json:"anonymous"`
		IdempotencyKey string             `json:"idempotencyKey"`
		Scheme         string             `json:"scheme"`
		RatingValue    *float64           `json:"ratingValue"`
		Quality        map[string]float64 `json:"quality"`
		MediaRefs      []string           `json:"mediaRefs"`
		Tags           []string           `json:"tags"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.AuthorID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.AuthorID = uid
		}
	}
	if key := r.Header.Get("Idempotency-Key"); key != "" && body.IdempotencyKey == "" {
		body.IdempotencyKey = key
	}
	res, err := h.Deps.CreateReview(r.Context(), tid, app.CreateReviewInput{
		AuthorID: body.AuthorID, TargetType: body.TargetType, TargetID: body.TargetID,
		OrderID: body.OrderID, Locale: body.Locale, Title: body.Title, Body: body.Body,
		Anonymous: body.Anonymous, IdempotencyKey: body.IdempotencyKey, Scheme: body.Scheme,
		RatingValue: body.RatingValue, Quality: body.Quality, MediaRefs: body.MediaRefs, Tags: body.Tags,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := map[string]any{"review": reviewDTO(res.Review), "moderation": map[string]any{
		"id": res.Case.ID.String(), "status": res.Case.Status, "autoDecision": res.Case.AutoDecision,
		"aiScore": res.Case.AIScore, "fraudScore": res.Case.FraudScore, "labels": res.Case.Labels,
	}}
	if res.Rating != nil {
		out["rating"] = map[string]any{
			"id": res.Rating.ID.String(), "scheme": res.Rating.Scheme,
			"value": res.Rating.Value, "stars": res.Rating.Stars, "verified": res.Rating.Verified,
		}
	}
	writeCreated(w, out)
}

func (h *Handler) getReview(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rev, err := h.Deps.Reviews.Get(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) updateReview(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Title    string    `json:"title"`
		Body     string    `json:"body"`
		AuthorID uuid.UUID `json:"authorId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	actor := body.AuthorID
	if actor == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			actor = uid
		}
	}
	rev, err := h.Deps.UpdateReview(r.Context(), tid, id, actor, body.Title, body.Body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) deleteReview(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	admin := r.URL.Query().Get("admin") == "true"
	actor, _ := UserIDFromContext(r.Context())
	if actor == uuid.Nil {
		actor, _ = uuid.Parse(r.URL.Query().Get("authorId"))
	}
	rev, err := h.Deps.DeleteReview(r.Context(), tid, id, actor, admin)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) listByTarget(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	targetType := r.PathValue("targetType")
	targetID, err := parsePathID(r, "targetId")
	if err != nil || !domain.ValidTarget(targetType) {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	list, err := h.Deps.Reviews.ListByTarget(r.Context(), tid, targetType, targetID, status, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, rev := range list {
		items = append(items, reviewDTO(rev))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) listByAuthor(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	authorID, err := parsePathID(r, "authorId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	list, err := h.Deps.Reviews.ListByAuthor(r.Context(), tid, authorID, 50)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, rev := range list {
		items = append(items, reviewDTO(rev))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) listRevisions(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	list, err := h.Deps.Reviews.ListRevisions(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) submitRating(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		AuthorID   uuid.UUID  `json:"authorId"`
		TargetType string     `json:"targetType"`
		TargetID   uuid.UUID  `json:"targetId"`
		OrderID    *uuid.UUID `json:"orderId"`
		Scheme     string     `json:"scheme"`
		Value      float64    `json:"value"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.AuthorID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.AuthorID = uid
		}
	}
	rt, ag, err := h.Deps.SubmitRating(r.Context(), tid, app.SubmitRatingInput{
		AuthorID: body.AuthorID, TargetType: body.TargetType, TargetID: body.TargetID,
		OrderID: body.OrderID, Scheme: body.Scheme, Value: body.Value,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"rating": map[string]any{
			"id": rt.ID.String(), "scheme": rt.Scheme, "value": rt.Value,
			"stars": rt.Stars, "verified": rt.Verified, "weight": rt.Weight,
		},
		"aggregate": aggregateDTO(ag),
	})
}

func aggregateDTO(ag domain.RatingAggregate) map[string]any {
	return map[string]any{
		"targetType": ag.TargetType, "targetId": ag.TargetID.String(), "scheme": ag.Scheme,
		"count": ag.Count, "avgStars": ag.AvgStars, "bayesianAvg": ag.BayesianAvg,
		"timeDecayAvg": ag.TimeDecayAvg, "verifiedCount": ag.VerifiedCount, "verifiedAvg": ag.VerifiedAvg,
		"updatedAt": ag.UpdatedAt,
	}
}

func (h *Handler) getAggregate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	targetType := r.PathValue("targetType")
	targetID, err := parsePathID(r, "targetId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	scheme := r.URL.Query().Get("scheme")
	if scheme == "" {
		scheme = domain.SchemeStars5
	}
	ag, err := h.Deps.Ratings.GetAggregate(r.Context(), tid, targetType, targetID, scheme)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, aggregateDTO(ag))
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		VoterID uuid.UUID `json:"voterId"`
		Helpful bool      `json:"helpful"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.VoterID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.VoterID = uid
		}
	}
	rev, err := h.Deps.VoteHelpful(r.Context(), tid, id, body.VoterID, body.Helpful)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) comment(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		AuthorID uuid.UUID  `json:"authorId"`
		Body     string     `json:"body"`
		ParentID *uuid.UUID `json:"parentId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.AuthorID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.AuthorID = uid
		}
	}
	c, err := h.Deps.AddComment(r.Context(), tid, id, body.AuthorID, body.Body, body.ParentID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": c.ID.String(), "reviewId": c.ReviewID.String(), "authorId": c.AuthorID.String(),
		"body": c.Body, "status": c.Status, "createdAt": c.CreatedAt,
	})
}

func (h *Handler) listComments(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	list, err := h.Deps.Interactions.ListComments(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ReporterID uuid.UUID `json:"reporterId"`
		Reason     string    `json:"reason"`
		Details    string    `json:"details"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.ReporterID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.ReporterID = uid
		}
	}
	rep, err := h.Deps.ReportReview(r.Context(), tid, id, body.ReporterID, body.Reason, body.Details)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"id": rep.ID.String(), "reason": rep.Reason})
}

func (h *Handler) pin(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rev, err := h.Deps.PinReview(r.Context(), tid, id, body.Pinned)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) moderationQueue(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	list, err := h.Deps.Moderation.ListPending(r.Context(), tid, 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list})
}

func (h *Handler) decideModeration(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	reviewID, err := parsePathID(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ModeratorID uuid.UUID `json:"moderatorId"`
		Approve     bool      `json:"approve"`
		Note        string    `json:"note"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.ModeratorID == uuid.Nil {
		if uid, ok := UserIDFromContext(r.Context()); ok {
			body.ModeratorID = uid
		}
	}
	rev, err := h.Deps.DecideModeration(r.Context(), tid, reviewID, body.ModeratorID, body.Approve, body.Note)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, reviewDTO(rev))
}

func (h *Handler) getTrust(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	reviewerID, err := parsePathID(r, "reviewerId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.Trust.Get(r.Context(), tid, reviewerID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"reviewerId": t.ReviewerID.String(), "score": t.Score, "badges": t.Badges,
		"verifiedPurchases": t.VerifiedPurchases, "publishedReviews": t.PublishedReviews,
		"rejectedReviews": t.RejectedReviews, "helpfulReceived": t.HelpfulReceived,
		"updatedAt": t.UpdatedAt,
	})
}

func (h *Handler) getReputation(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	targetType := r.PathValue("targetType")
	targetID, err := parsePathID(r, "targetId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rep, err := h.Deps.Reputation.Get(r.Context(), tid, targetType, targetID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"targetType": rep.TargetType, "targetId": rep.TargetID.String(),
		"score": rep.Score, "tier": rep.Tier, "reviewCount": rep.ReviewCount, "updatedAt": rep.UpdatedAt,
	})
}

func (h *Handler) summarize(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	targetType := r.PathValue("targetType")
	targetID, err := parsePathID(r, "targetId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	sum, err := h.Deps.SummarizeTarget(r.Context(), tid, targetType, targetID, 20)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"summary": sum})
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	targetType := r.URL.Query().Get("targetType")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if h.Deps.Search == nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	ids, err := h.Deps.Search.Search(r.Context(), tid, q, targetType, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		if rev, err := h.Deps.Reviews.Get(r.Context(), tid, id); err == nil {
			items = append(items, reviewDTO(rev))
		}
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) adminStats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	stats, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, stats)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
