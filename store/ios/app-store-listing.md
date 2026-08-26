# App Store listing — NEXORA Customer

Bundle ID: `com.hfzx.nexora.nexoraCustomer`
Version / build: from `apps/mobile_customer/pubspec.yaml` (`1.0.0+1`)
Deployment target: Flutter default iOS (see Xcode project)

## Copy (draft — not submitted)

| Field | Value |
|---|---|
| Name | NEXORA |
| Subtitle | Quick commerce at your door |
| Description | Browse nearby stores, checkout in minutes, and track your courier. |
| Keywords | grocery,delivery,quick commerce |
| Support URL | **EXTERNAL INPUT REQUIRED** |
| Marketing URL | optional — **EXTERNAL INPUT REQUIRED** |
| Age rating | **EXTERNAL INPUT REQUIRED** (complete Apple questionnaire) |
| Review notes | Demo tenant and sandbox payment adapter; no live card capture in CI. |
| Release notes (1.0.0) | Initial production candidate. |

## Assets required (not in this repository)

- iPhone 6.7" / 6.5" / 5.5" screenshot sets
- iPad screenshots if iPad listed
- Privacy nutrition labels (App Privacy)

## Signing / Apple — EXTERNAL INPUTS REQUIRED

- Apple Developer team + certificates (do not commit)
- Provisioning profiles
- `aps-environment` entitlement when push is enabled in production
- Associated domains / universal links host ownership for `nexora.io`
- App Store Connect app record

CI `rc-ios-build` compiles `flutter build ios --release --no-codesign`. Codesign, Notarization, and ASC upload remain **BLOCKED** without org secrets.
