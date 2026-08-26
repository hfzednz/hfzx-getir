# Play Store listing — NEXORA Customer

Application ID: `com.hfzx.nexora.nexora_customer`
Version name / code: from `apps/mobile_customer/pubspec.yaml` (`1.0.0+1`)

## Copy (draft — not submitted)

| Field | Value |
|---|---|
| Application name | NEXORA |
| Short description | Quick commerce delivery in minutes. |
| Full description | NEXORA is a quick-commerce customer app: browse nearby dark-store inventory, add to cart, checkout, and track delivery. |
| Release notes (1.0.0) | Initial production candidate. |

## Assets required (not in this repository)

- Phone screenshots (min 2, max 8) — 16:9 or 9:16
- 7-inch / 10-inch tablet screenshots if tablet-listed
- Feature graphic 1024×500
- High-res icon 512×512 (Play Console, not the mipmap in-app icon)

## Policy / legal — EXTERNAL INPUTS REQUIRED

Do not invent these. Provide before store submission:

- Privacy policy URL
- Data safety form answers (location, photos/camera, files, personal info)
- Content rating questionnaire
- Target audience / age
- Developer account identity (Google Play Console)
- Production signing keystore (CI env: `ANDROID_KEYSTORE_PATH`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`)
- Play service account JSON (`GOOGLE_PLAY_SERVICE_ACCOUNT_JSON`) for Fastlane

## Permissions justification

| Permission | Why |
|---|---|
| INTERNET | API / images |
| ACCESS_FINE/COARSE_LOCATION | Delivery ETA and serviceability |
| CAMERA | Barcode / visual search |
| RECORD_AUDIO | Voice search |
| POST_NOTIFICATIONS | Order status |

CI `rc-android-aab` produces an AAB. Without Play Console + upload key it is **not** store-published.
