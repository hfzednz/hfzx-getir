package app

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
)

func defaultDestPhoneRoles() map[string][]string {
	return map[string][]string{
		"+905551112233": {"customer"},
		"+905551112234": {"picker", "packer", "dispatcher"},
		"+905551112235": {"courier"},
		"+905551112236": {"finance_analyst"},
		"+905551112237": {"support_agent"},
		"+905551112238": {"city_ops"},
		"+905551112239": {"admin"},
		"+905551112240": {"supplier", "partner"},
		"+905551112241": {"super_admin"},
	}
}

func parseDevPhoneRoles(raw string) map[string][]string {
	out := map[string][]string{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		phone, roles, ok := strings.Cut(part, "=")
		if !ok {
			phone, roles, ok = strings.Cut(part, ":")
			if !ok {
				continue
			}
		}
		phone = domain.NormalizeIdentifier(domain.IdentifierTypePhone, strings.TrimSpace(phone))
		var list []string
		for _, r := range strings.Split(roles, ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				list = append(list, r)
			}
		}
		if phone != "" && len(list) > 0 {
			out[phone] = list
		}
	}
	return out
}

func otpRolesForPhone(phone string) []string {
	phone = domain.NormalizeIdentifier(domain.IdentifierTypePhone, phone)
	if os.Getenv("OTP_DEV_MODE") != "true" {
		return nil
	}
	m := parseDevPhoneRoles(os.Getenv("IDENTITY_DEV_PHONE_ROLES"))
	if len(m) == 0 {
		m = defaultDestPhoneRoles()
	}
	return m[phone]
}

func (d *Deps) ensureOTPRoles(ctx context.Context, principalID uuid.UUID, phone string) error {
	if d.Roles == nil {
		return nil
	}
	existing, err := d.roleNames(ctx, principalID)
	if err != nil {
		return err
	}
	have := map[string]struct{}{}
	for _, n := range existing {
		have[n] = struct{}{}
	}
	wanted := otpRolesForPhone(phone)
	if len(wanted) == 0 && len(existing) > 0 {
		return nil
	}
	if len(wanted) == 0 {
		wanted = []string{"customer"}
	}
	now := d.now()
	for _, name := range wanted {
		if _, ok := have[name]; ok {
			continue
		}
		role, err := d.Roles.GetRoleByName(ctx, nil, name)
		if err != nil {
			continue
		}
		if err := d.Roles.AssignRole(ctx, domain.PrincipalRole{
			ID: d.newID(), PrincipalID: principalID, RoleID: role.ID, CreatedAt: now,
		}); err != nil {
			return err
		}
	}
	return nil
}
