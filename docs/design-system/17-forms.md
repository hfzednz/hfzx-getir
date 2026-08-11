# 17 — Forms

## Anatomy

Label → Field → Helper / Error → Optional trailing hint  
Gaps: 16 between fields; 8 label-to-field; 4 field-to-error.

## Validation

| Timing | Rule |
|--------|------|
| On blur | Format validation |
| On submit | Full form validation; focus first error |
| On change | Clear error when corrected; live for OTP length |

Error text: specific, actionable (`Enter a valid TR mobile number`) — not `Invalid`.

## Input masks

| Field | Mask / behavior |
|-------|-----------------|
| Phone | Locale national format; store E.164 |
| OTP | Digits only; length 6 default; autofill |
| PIN | Obscured dots; length 4–6 |
| Email | Trim; lowercase on submit |
| Password | Show/hide toggle; strength meter optional |
| Credit card | PSP fields preferred (hosted); if native: groups 4s, Luhn helper |
| Expiry | MM/YY |
| CVV | 3–4 obscured |
| Address | Line clamps; apartment optional marked |
| Coupon | Uppercase alnum; apply button |
| Gift card | Grouped code; balance check async |

Money inputs: decimal per currency; never float drift — decimal string → minor units at boundary.

## OTP — `NxOtpInput`

- Auto-advance; backspace retreats
- Paste fills all
- Resend with cooldown timer (tabular)
- Error: clear values or highlight all

## Accessibility

- Labels associated (`for` / `Semantics`)
- Errors linked via `aria-describedby` / Semantics
- Required indicated in text, not color alone
- Autocomplete attributes on web

## Admin forms

- Dense; inline validation
- Unsaved changes guard
- Keyboard: Enter submits primary when safe
