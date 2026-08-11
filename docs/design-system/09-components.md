# 09 — Component Library

> Spec contracts for `nexora_design` / `@nexora/ui`.  
> Naming: `Nx` prefix. Presentational only — callbacks in, no repositories.

## Inventory & contracts

### Buttons — `NxButton`

| Variant | Visual | Use |
|---------|--------|-----|
| `primary` | `bg.accent` + `text.onAccent` (customer) OR `bg.brand` + `text.onBrand` (ops) | Highest priority |
| `secondary` | surface + border.default | Secondary |
| `tertiary` | text.brand, no fill | Low emphasis |
| `danger` | danger fill / outline | Destructive |
| `ghost` | transparent | Toolbars |

Sizes: `sm` 32h · `md` 44h · `lg` 52h  
States: default, hover, pressed, focus, loading, disabled  
Loading: replace label with `NxSpinner` size sm; keep width stable  
Full-width option for mobile CTAs

### Icon Buttons — `NxIconButton`

Sizes 32 / 40 / 48 hit area; variants ghost / secondary / filled. Tooltip on desktop.

### FAB — `NxFab`

Elevation 3; brand or accent; extended FAB label optional (admin rare). One FAB per screen max.

### Cards — `NxCard`

Padding tokenized; radius md; elevation 0–1. **Use only when container aids interaction/understanding.** Prefer flat list rows when chrome adds nothing.

Specialized: see §§10–11 for product/order cards.

### Dialogs — `NxDialog`

Max width 400 mobile / 480 desktop; title + body + actions (max 2 primary-row); destructive confirms require typing or explicit checkbox when high risk (admin).

### Bottom Sheets — `NxSheet`

Snap points 35% / 55% / 90%; drag handle; radius.lg top; keyboard inset aware. Prefer sheets over dialogs on mobile for filters, cart, substitutes.

### Snackbars — `NxToast`

Bottom above nav/cart bar; 1 action max; auto-dismiss 3–5s; variants info/success/warning/danger; do not stack >2.

### Inputs — `NxTextField`

Label outside or floating (prefer outside for forms density); helper; error text; leading/trailing icons; sm radius.

### Search Bars — `NxSearchField`

Leading search icon; clear; optional voice/scan trailing; cancels to previous route on mobile search page.

### Dropdowns — `NxSelect` / `NxMenu`

Desktop popover elevation 2; mobile may open sheet. Virtualize long lists.

### Checkboxes / Radios / Switches — `NxCheckbox` `NxRadio` `NxSwitch`

Brand selected fill; label hit target includes text; group semantics.

### Badges / Chips / Tags — `NxBadge` `NxChip` `NxTag`

| Type | Use |
|------|-----|
| Badge | Counts, status dots |
| Chip | Filters, selectable |
| Tag | Read-only meta (discount, dietary) |

### Tabs — `NxTabs`

Underline indicator brand; scrollable on mobile; equal-width rare.

### Accordions — `NxAccordion`

For PDP nutrition / FAQ — not for primary checkout path.

### Progress — `NxProgress` `NxStepper`

Linear for uploads; stepper for checkout / warehouse pack stages.

### Sliders — `NxSlider`

Rare (tips). Large thumb 24; value label tabular.

### Date / Time / Calendar — `NxDatePicker` `NxTimePicker` `NxCalendar`

City timezone labeled; locale calendars; admin range picker.

### Ratings — `NxRating`

Stars 24; half-star if data supports; read-only vs input modes.

### Avatar — `NxAvatar`

Sizes 24/32/40/56; image / initials / icon; status ring optional (courier online).

### Profile Cards — `NxProfileHeader`

Avatar + name + meta; actions overflow.

### Carousels — `NxCarousel`

Page dots; peek next card 12–16; autoplay **off** by default; pause on focus.

### Banners — `NxBanner`

Inline page banner (info/warn); dismissible; not sticker-on-hero.

### Navigation chrome

| Component | Surfaces |
|-----------|----------|
| `NxBottomNav` | Customer, Courier |
| `NxNavDrawer` | Rare mobile admin |
| `NxNavRail` | Tablet |
| `NxSidebar` | Desktop admin / super admin |
| `NxTopBar` | All — title, actions, city/store context |

### Data Tables — `NxDataTable`

Sticky header; tabular cells; row hover desktop; bulk select; empty/error states.

### Charts — `NxChart`

Tokenized series colors (teal family + neutrals + 1 citrus highlight). No rainbow defaults. Tooltip accessible.

### Timeline — `NxTimeline`

Order tracking vertical; past/current/future states.

### Stepper — `NxStepper`

Horizontal checkout; vertical warehouse.

### OTP / PIN — `NxOtpInput` `NxPinInput`

Equal boxes; autofill; paste support; error shake micro (reduced-motion: highlight border).

### Address Picker — `NxAddressPicker`

List + map pin refine; accuracy banner; serviceability chip.

### Map — `NxMapChrome`

See [19-map.md](19-map.md) — controls, legend, courier marker, ETA chip.

---

## State matrix (all interactive)

`default | hover | pressed | focusVisible | disabled | loading | error`

Document per Storybook/Widgetbook story.

## Do / Don't

- DO compose patterns from primitives  
- DON'T nest cards in cards  
- DON'T invent third primary button color  
- DON'T use Material `ElevatedButton` raw in apps
