package httpadapter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app"
	"github.com/nexora/customer-profile-service/internal/domain"
	"github.com/nexora/customer-profile-service/internal/observability"
)

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]any{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]any{"status": "ready"})
}

func (h *Handler) pathID(r *http.Request, name string) (uuid.UUID, error) {
	raw := r.PathValue(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid %s", domain.ErrInvalidArgument, name)
	}
	return id, nil
}

func (h *Handler) requirePrincipal(r *http.Request) (TrustedPrincipal, error) {
	p, ok := PrincipalFromContext(r.Context())
	if !ok || p.PrincipalID == uuid.Nil {
		return TrustedPrincipal{}, domain.ErrUnauthorized
	}
	if p.TenantID == uuid.Nil {
		return TrustedPrincipal{}, fmt.Errorf("%w: X-Nexora-Tenant / X-Tenant-Id required", domain.ErrUnauthorized)
	}
	return p, nil
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	tp, err := h.requirePrincipal(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetProfileByPrincipal(r.Context(), tp.TenantID, tp.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

func (h *Handler) patchMe(w http.ResponseWriter, r *http.Request) {
	tp, err := h.requirePrincipal(r)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetProfileByPrincipal(r.Context(), tp.TenantID, tp.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body patchProfileBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	updated, err := h.Deps.UpdateProfile(r.Context(), body.toInput(p.ID, RequestIDFromContext(r.Context())))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(updated))
}

type provisionBody struct {
	TenantID    string `json:"tenantId"`
	PrincipalID string `json:"principalId"`
	DisplayName string `json:"displayName"`
	FullName    string `json:"fullName"`
	Nickname    string `json:"nickname"`
	Language    string `json:"language"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
}

func (h *Handler) provisionCustomer(w http.ResponseWriter, r *http.Request) {
	var body provisionBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	tenantID, err := uuid.Parse(body.TenantID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: tenantId", domain.ErrInvalidArgument))
		return
	}
	principalID, err := uuid.Parse(body.PrincipalID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: principalId", domain.ErrInvalidArgument))
		return
	}
	p, err := h.Deps.Provision(r.Context(), app.ProvisionInput{
		TenantID: tenantID, PrincipalID: principalID,
		DisplayName: body.DisplayName, FullName: body.FullName, Nickname: body.Nickname,
		Language: body.Language, CountryCode: body.CountryCode, City: body.City, Timezone: body.Timezone,
		TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	observability.Default.ProfilesCreated.Add(1)
	writeCreated(w, profileDTO(p))
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetProfile(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

type patchProfileBody struct {
	DisplayName   *string         `json:"displayName"`
	FullName      *string         `json:"fullName"`
	Nickname      *string         `json:"nickname"`
	Gender        *string         `json:"gender"`
	Language      *string         `json:"language"`
	CountryCode   *string         `json:"countryCode"`
	City          *string         `json:"city"`
	Timezone      *string         `json:"timezone"`
	Occupation    *string         `json:"occupation"`
	FamilySize    *int            `json:"familySize"`
	Dietary       map[string]any  `json:"dietary"`
	Accessibility map[string]any  `json:"accessibility"`
}

func (b patchProfileBody) toInput(profileID uuid.UUID, traceID string) app.UpdateProfileInput {
	in := app.UpdateProfileInput{
		ProfileID: profileID, DisplayName: b.DisplayName, FullName: b.FullName, Nickname: b.Nickname,
		Language: b.Language, CountryCode: b.CountryCode, City: b.City, Timezone: b.Timezone,
		Occupation: b.Occupation, FamilySize: b.FamilySize, Dietary: b.Dietary, Accessibility: b.Accessibility,
		TraceID: traceID,
	}
	if b.Gender != nil {
		g := domain.Gender(*b.Gender)
		in.Gender = &g
	}
	return in
}

func (h *Handler) patchCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body patchProfileBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.UpdateProfile(r.Context(), body.toInput(id, RequestIDFromContext(r.Context())))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

func (h *Handler) listAddresses(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	addrs, err := h.Deps.ListAddresses(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(addrs))
	for _, a := range addrs {
		items = append(items, addressDTO(a))
	}
	writeOK(w, map[string]any{"items": items})
}

type addressBody struct {
	Label       string  `json:"label"`
	CustomLabel string  `json:"customLabel"`
	Line1       string  `json:"line1"`
	Building    string  `json:"building"`
	Apartment   string  `json:"apartment"`
	Entrance    string  `json:"entrance"`
	Floor       string  `json:"floor"`
	Door        string  `json:"door"`
	Notes       string  `json:"notes"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	IsDefault   bool    `json:"isDefault"`
}

func (h *Handler) createAddress(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body addressBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	a, err := h.Deps.AddAddress(r.Context(), app.AddAddressInput{
		ProfileID: id, Label: domain.AddressLabel(body.Label), CustomLabel: body.CustomLabel,
		Line1: body.Line1, Building: body.Building, Apartment: body.Apartment, Entrance: body.Entrance,
		Floor: body.Floor, Door: body.Door, Notes: body.Notes, Lat: body.Lat, Lng: body.Lng,
		IsDefault: body.IsDefault, TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, addressDTO(a))
}

type patchAddressBody struct {
	Label       *string  `json:"label"`
	CustomLabel *string  `json:"customLabel"`
	Line1       *string  `json:"line1"`
	Building    *string  `json:"building"`
	Apartment   *string  `json:"apartment"`
	Entrance    *string  `json:"entrance"`
	Floor       *string  `json:"floor"`
	Door        *string  `json:"door"`
	Notes       *string  `json:"notes"`
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
}

func (h *Handler) patchAddress(w http.ResponseWriter, r *http.Request) {
	addrID, err := h.pathID(r, "addressId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body patchAddressBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	in := app.UpdateAddressInput{
		AddressID: addrID, CustomLabel: body.CustomLabel, Line1: body.Line1, Building: body.Building,
		Apartment: body.Apartment, Entrance: body.Entrance, Floor: body.Floor, Door: body.Door,
		Notes: body.Notes, Lat: body.Lat, Lng: body.Lng, TraceID: RequestIDFromContext(r.Context()),
	}
	if body.Label != nil {
		l := domain.AddressLabel(*body.Label)
		in.Label = &l
	}
	a, err := h.Deps.UpdateAddress(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, addressDTO(a))
}

func (h *Handler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	addrID, err := h.pathID(r, "addressId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.DeleteAddress(r.Context(), addrID, RequestIDFromContext(r.Context())); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) setDefaultAddress(w http.ResponseWriter, r *http.Request) {
	profileID, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	addrID, err := h.pathID(r, "addressId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	a, err := h.Deps.SetDefaultAddress(r.Context(), profileID, addrID, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, addressDTO(a))
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetPreferences(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, prefsDTO(p))
}

type prefsBody struct {
	FavoriteBrands     []string       `json:"favoriteBrands"`
	FavoriteCategories []string       `json:"favoriteCategories"`
	FavoriteProducts   []string       `json:"favoriteProducts"`
	FavoriteStores     []string       `json:"favoriteStores"`
	Delivery           map[string]any `json:"delivery"`
	Payment            map[string]any `json:"payment"`
	Notification       map[string]any `json:"notification"`
	Shopping           map[string]any `json:"shopping"`
	Theme              string         `json:"theme"`
	Language           string         `json:"language"`
	Accessibility      map[string]any `json:"accessibility"`
}

func parseUUIDList(ss []string) ([]uuid.UUID, error) {
	if ss == nil {
		return nil, nil
	}
	out := make([]uuid.UUID, 0, len(ss))
	for _, s := range ss {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid uuid %q", domain.ErrInvalidArgument, s)
		}
		out = append(out, id)
	}
	return out, nil
}

func (h *Handler) putPreferences(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body prefsBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	brands, err := parseUUIDList(body.FavoriteBrands)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	cats, err := parseUUIDList(body.FavoriteCategories)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	prods, err := parseUUIDList(body.FavoriteProducts)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	stores, err := parseUUIDList(body.FavoriteStores)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.UpsertPreferences(r.Context(), app.UpsertPreferencesInput{
		ProfileID: id, FavoriteBrands: brands, FavoriteCategories: cats,
		FavoriteProducts: prods, FavoriteStores: stores,
		Delivery: body.Delivery, Payment: body.Payment, Notification: body.Notification,
		Shopping: body.Shopping, Theme: body.Theme, Language: body.Language,
		Accessibility: body.Accessibility, TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, prefsDTO(p))
}

func (h *Handler) setAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	ct := r.Header.Get("Content-Type")
	filename := "avatar.bin"
	var body io.Reader = r.Body
	if strings.HasPrefix(ct, "multipart/") {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			writeErr(w, r, fmt.Errorf("%w: %v", domain.ErrInvalidArgument, err))
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			writeErr(w, r, fmt.Errorf("%w: file required", domain.ErrInvalidArgument))
			return
		}
		defer file.Close()
		filename = hdr.Filename
		ct = hdr.Header.Get("Content-Type")
		body = file
	} else {
		defer r.Body.Close()
		buf, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
		if err != nil {
			writeErr(w, r, err)
			return
		}
		body = bytes.NewReader(buf)
	}
	p, err := h.Deps.SetAvatar(r.Context(), app.SetAvatarInput{
		ProfileID: id, Filename: filename, ContentType: ct, Body: body,
		TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

func (h *Handler) deleteAvatar(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.DeleteAvatar(r.Context(), id, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	tags, err := h.Deps.ListTags(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		items = append(items, tagAssignDTO(t))
	}
	writeOK(w, map[string]any{"items": items})
}

type addTagBody struct {
	TagID string `json:"tagId"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Note  string `json:"note"`
}

func (h *Handler) addTag(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body addTagBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	var tagID uuid.UUID
	if body.TagID != "" {
		tagID, err = uuid.Parse(body.TagID)
		if err != nil {
			writeErr(w, r, fmt.Errorf("%w: tagId", domain.ErrInvalidArgument))
			return
		}
	}
	pt, err := h.Deps.AddTag(r.Context(), app.AddTagInput{
		ProfileID: id, TagID: tagID, Kind: domain.TagKind(body.Kind), Name: body.Name, Note: body.Note,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, tagAssignDTO(pt))
}

func (h *Handler) removeTag(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	tagID, err := h.pathID(r, "tagId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.RemoveTag(r.Context(), id, tagID); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) getHousehold(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	hh, err := h.Deps.GetHouseholdByOwner(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	members, _ := h.Deps.ListHouseholdMembers(r.Context(), hh.ID)
	memberDTOs := make([]map[string]any, 0, len(members))
	for _, m := range members {
		memberDTOs = append(memberDTOs, map[string]any{
			"id": m.ID, "householdId": m.HouseholdID, "profileId": m.ProfileID,
			"role": m.Role, "joinedAt": m.JoinedAt,
		})
	}
	out := householdDTO(hh)
	out["members"] = memberDTOs
	writeOK(w, out)
}

type createHouseholdBody struct {
	Name    string `json:"name"`
	Sharing struct {
		Addresses bool `json:"addresses"`
		Payments  bool `json:"payments"`
		Lists     bool `json:"lists"`
		Wallet    bool `json:"wallet"`
		Loyalty   bool `json:"loyalty"`
	} `json:"sharing"`
}

func (h *Handler) createHousehold(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body createHouseholdBody
	_ = decodeJSON(r, &body)
	hh, err := h.Deps.CreateHousehold(r.Context(), app.CreateHouseholdInput{
		OwnerProfileID: id, Name: body.Name,
		Sharing: domain.HouseholdSharingFlags{
			Addresses: body.Sharing.Addresses, Payments: body.Sharing.Payments,
			Lists: body.Sharing.Lists, Wallet: body.Sharing.Wallet, Loyalty: body.Sharing.Loyalty,
		},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, householdDTO(hh))
}

type addMemberBody struct {
	ProfileID string `json:"profileId"`
	Role      string `json:"role"`
}

func (h *Handler) addHouseholdMember(w http.ResponseWriter, r *http.Request) {
	ownerID, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	hh, err := h.Deps.GetHouseholdByOwner(r.Context(), ownerID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body addMemberBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	pid, err := uuid.Parse(body.ProfileID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: profileId", domain.ErrInvalidArgument))
		return
	}
	m, err := h.Deps.AddHouseholdMember(r.Context(), app.AddHouseholdMemberInput{
		HouseholdID: hh.ID, ProfileID: pid, Role: domain.HouseholdMemberRole(body.Role),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": m.ID, "householdId": m.HouseholdID, "profileId": m.ProfileID,
		"role": m.Role, "joinedAt": m.JoinedAt,
	})
}

func (h *Handler) updateHouseholdSharing(w http.ResponseWriter, r *http.Request) {
	ownerID, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	hh, err := h.Deps.GetHouseholdByOwner(r.Context(), ownerID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body createHouseholdBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	updated, err := h.Deps.UpdateSharing(r.Context(), app.UpdateSharingInput{
		HouseholdID: hh.ID,
		Sharing: domain.HouseholdSharingFlags{
			Addresses: body.Sharing.Addresses, Payments: body.Sharing.Payments,
			Lists: body.Sharing.Lists, Wallet: body.Sharing.Wallet, Loyalty: body.Sharing.Loyalty,
		},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, householdDTO(updated))
}

func (h *Handler) listConsents(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	list, err := h.Deps.ListConsents(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, c := range list {
		items = append(items, consentDTO(c))
	}
	writeOK(w, map[string]any{"items": items})
}

type consentBody struct {
	Channel string `json:"channel"`
	Granted bool   `json:"granted"`
	Source  string `json:"source"`
}

func (h *Handler) setConsent(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body consentBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	c, err := h.Deps.SetConsent(r.Context(), app.SetConsentInput{
		ProfileID: id, Channel: domain.ConsentChannel(body.Channel),
		Granted: body.Granted, Source: body.Source, TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, consentDTO(c))
}

func (h *Handler) getCustomer360(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	view, err := h.Deps.GetCustomer360(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := map[string]any{"profile": profileDTO(view.Profile)}
	if view.Preferences != nil {
		out["preferences"] = prefsDTO(*view.Preferences)
	}
	addrs := make([]map[string]any, 0, len(view.Addresses))
	for _, a := range view.Addresses {
		addrs = append(addrs, addressDTO(a))
	}
	out["addresses"] = addrs
	tags := make([]map[string]any, 0, len(view.Tags))
	for _, t := range view.Tags {
		tags = append(tags, tagAssignDTO(t))
	}
	out["tags"] = tags
	consents := make([]map[string]any, 0, len(view.Consents))
	for _, c := range view.Consents {
		consents = append(consents, consentDTO(c))
	}
	out["consents"] = consents
	notes := make([]map[string]any, 0, len(view.Notes))
	for _, n := range view.Notes {
		notes = append(notes, noteDTO(n))
	}
	out["notes"] = notes
	tl := make([]map[string]any, 0, len(view.Timeline))
	for _, e := range view.Timeline {
		tl = append(tl, timelineDTO(e))
	}
	out["timeline"] = tl
	segs := make([]map[string]any, 0, len(view.Segments))
	for _, s := range view.Segments {
		segs = append(segs, membershipDTO(s))
	}
	out["segments"] = segs
	if view.Personalization != nil {
		out["personalization"] = personalizationDTO(*view.Personalization)
	}
	if view.AIModel != nil {
		out["aiModel"] = aiModelDTO(*view.AIModel)
	}
	if view.Household != nil {
		out["household"] = householdDTO(*view.Household)
	}
	writeOK(w, out)
}

func (h *Handler) listNotes(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	notes, err := h.Deps.ListNotes(r.Context(), id, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		items = append(items, noteDTO(n))
	}
	writeOK(w, map[string]any{"items": items})
}

type noteBody struct {
	AuthorID string `json:"authorId"`
	Body     string `json:"body"`
	Pinned   bool   `json:"pinned"`
}

func (h *Handler) addNote(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body noteBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	authorID, err := uuid.Parse(body.AuthorID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: authorId", domain.ErrInvalidArgument))
		return
	}
	n, err := h.Deps.AddNote(r.Context(), app.AddNoteInput{
		ProfileID: id, AuthorID: authorID, Body: body.Body, Pinned: body.Pinned,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, noteDTO(n))
}

func (h *Handler) listTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.Deps.ListTimeline(r.Context(), id, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(events))
	for _, e := range events {
		items = append(items, timelineDTO(e))
	}
	writeOK(w, map[string]any{"items": items})
}

type timelineBody struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
	ActorID string         `json:"actorId"`
}

func (h *Handler) appendTimeline(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body timelineBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	in := app.AppendTimelineInput{ProfileID: id, Type: body.Type, Payload: body.Payload}
	if body.ActorID != "" {
		aid, err := uuid.Parse(body.ActorID)
		if err != nil {
			writeErr(w, r, fmt.Errorf("%w: actorId", domain.ErrInvalidArgument))
			return
		}
		in.ActorID = &aid
	}
	e, err := h.Deps.AppendTimeline(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, timelineDTO(e))
}

func (h *Handler) listSegments(w http.ResponseWriter, r *http.Request) {
	tenantRaw := r.URL.Query().Get("tenantId")
	if tenantRaw == "" {
		tenantRaw = r.Header.Get("X-Nexora-Tenant")
		if tenantRaw == "" {
			tenantRaw = r.Header.Get("X-Tenant-Id")
		}
	}
	tenantID, err := uuid.Parse(tenantRaw)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: tenantId required", domain.ErrInvalidArgument))
		return
	}
	segs, err := h.Deps.ListSegments(r.Context(), tenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(segs))
	for _, s := range segs {
		items = append(items, segmentDTO(s))
	}
	writeOK(w, map[string]any{"items": items})
}

type segmentBody struct {
	ID          string         `json:"id"`
	TenantID    string         `json:"tenantId"`
	Name        string         `json:"name"`
	Kind        string         `json:"kind"`
	Description string         `json:"description"`
	Rules       map[string]any `json:"rules"`
	Active      bool           `json:"active"`
}

func (h *Handler) upsertSegment(w http.ResponseWriter, r *http.Request) {
	var body segmentBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	tenantID, err := uuid.Parse(body.TenantID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: tenantId", domain.ErrInvalidArgument))
		return
	}
	s := domain.Segment{
		TenantID: tenantID, Name: body.Name, Kind: domain.SegmentKind(body.Kind),
		Description: body.Description, Rules: body.Rules, Active: body.Active,
	}
	if body.ID != "" {
		s.ID, err = uuid.Parse(body.ID)
		if err != nil {
			writeErr(w, r, fmt.Errorf("%w: id", domain.ErrInvalidArgument))
			return
		}
	}
	out, err := h.Deps.UpsertSegmentDefinition(r.Context(), s)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, segmentDTO(out))
}

func (h *Handler) listProfileSegments(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	list, err := h.Deps.ListProfileSegments(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, m := range list {
		items = append(items, membershipDTO(m))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) assignSegment(w http.ResponseWriter, r *http.Request) {
	profileID, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	segmentID, err := h.pathID(r, "segmentId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	m, err := h.Deps.AssignSegment(r.Context(), app.AssignSegmentInput{
		SegmentID: segmentID, ProfileID: profileID, Source: "admin",
		TraceID: RequestIDFromContext(r.Context()),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, membershipDTO(m))
}

func (h *Handler) evaluateSegment(w http.ResponseWriter, r *http.Request) {
	profileID, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	segmentID, err := h.pathID(r, "segmentId")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	matched, err := h.Deps.EvaluateDynamic(r.Context(), segmentID, profileID, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"matched": matched})
}

func (h *Handler) getPersonalization(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.GetPersonalization(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, personalizationDTO(p))
}

type personalizationBody struct {
	Homepage       map[string]any `json:"homepage"`
	Category       map[string]any `json:"category"`
	Recommendation map[string]any `json:"recommendation"`
	Search         map[string]any `json:"search"`
	Delivery       map[string]any `json:"delivery"`
	Promotion      map[string]any `json:"promotion"`
	ShoppingHabits map[string]any `json:"shoppingHabits"`
}

func (h *Handler) putPersonalization(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body personalizationBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.UpdatePersonalization(r.Context(), app.UpdatePersonalizationInput{
		ProfileID: id, Homepage: body.Homepage, Category: body.Category,
		Recommendation: body.Recommendation, Search: body.Search, Delivery: body.Delivery,
		Promotion: body.Promotion, ShoppingHabits: body.ShoppingHabits,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, personalizationDTO(p))
}

func (h *Handler) getAIModel(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	m, err := h.Deps.GetAIModel(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, aiModelDTO(m))
}

type recomputeBody struct {
	OrderCount30d int      `json:"orderCount30d"`
	OrderCount90d int      `json:"orderCount90d"`
	AvgOrderValue float64  `json:"avgOrderValue"`
	DaysSinceLast int      `json:"daysSinceLast"`
	CancelRate    float64  `json:"cancelRate"`
	PreferredCats []string `json:"preferredCats"`
}

func (h *Handler) recomputeAIModel(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body recomputeBody
	_ = decodeJSON(r, &body)
	m, err := h.Deps.RecomputeAIModel(r.Context(), id, domain.ActivityOrdersSummary{
		OrderCount30d: body.OrderCount30d, OrderCount90d: body.OrderCount90d,
		AvgOrderValue: body.AvgOrderValue, DaysSinceLast: body.DaysSinceLast,
		CancelRate: body.CancelRate, PreferredCats: body.PreferredCats,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, aiModelDTO(m))
}

func (h *Handler) requestExport(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	req, err := h.Deps.RequestExport(r.Context(), id, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, privacyDTO(req))
}

func (h *Handler) requestDeletion(w http.ResponseWriter, r *http.Request) {
	id, err := h.pathID(r, "id")
	if err != nil {
		writeErr(w, r, err)
		return
	}
	req, err := h.Deps.RequestDeletion(r.Context(), id, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, privacyDTO(req))
}

func (h *Handler) adminSearch(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(r.URL.Query().Get("tenantId"))
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: tenantId required", domain.ErrInvalidArgument))
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	q := r.URL.Query().Get("q")
	list, err := h.Deps.SearchCustomers(r.Context(), tenantID, q, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, p := range list {
		items = append(items, profileDTO(p))
	}
	writeOK(w, map[string]any{"items": items})
}

type mergeBody struct {
	TargetID string `json:"targetId"`
	SourceID string `json:"sourceId"`
}

func (h *Handler) adminMerge(w http.ResponseWriter, r *http.Request) {
	var body mergeBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, err)
		return
	}
	targetID, err := uuid.Parse(body.TargetID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: targetId", domain.ErrInvalidArgument))
		return
	}
	sourceID, err := uuid.Parse(body.SourceID)
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: sourceId", domain.ErrInvalidArgument))
		return
	}
	p, err := h.Deps.MergeCustomers(r.Context(), targetID, sourceID, RequestIDFromContext(r.Context()))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, profileDTO(p))
}

func (h *Handler) adminDuplicates(w http.ResponseWriter, r *http.Request) {
	tenantID, err := uuid.Parse(r.URL.Query().Get("tenantId"))
	if err != nil {
		writeErr(w, r, fmt.Errorf("%w: tenantId required", domain.ErrInvalidArgument))
		return
	}
	list, err := h.Deps.FindDuplicates(r.Context(), tenantID, r.URL.Query().Get("displayName"), r.URL.Query().Get("fullName"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, p := range list {
		items = append(items, profileDTO(p))
	}
	writeOK(w, map[string]any{"items": items})
}
