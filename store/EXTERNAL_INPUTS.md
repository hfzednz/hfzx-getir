# Store / legal inputs that this repository does not contain

Fill these outside git. Do not commit secrets, certificates, or invented legal URLs.

| Input | Android | iOS | Status |
|---|---|---|---|
| Privacy policy URL | required | required | missing |
| Support URL | Play Console | App Store Connect | missing |
| Marketing URL | optional | optional | missing |
| Company legal name / DUNS | Play + Apple accounts | Apple | missing |
| Content / age rating answers | Play questionnaire | Apple questionnaire | missing |
| Data safety / nutrition labels | Play Data safety | App Privacy | missing |
| Upload keystore / Play App Signing | env secrets listed in `cd-mobile.yml` | — | not in repo (correct) |
| Apple certs / profiles / ASC API key | — | org secrets | not in repo (correct) |
| Production `GOOGLE_MAPS_API_KEY` | Android manifest placeholder | iOS if maps used | placeholder only |
| Firebase / FCM production apps | google-services / GoogleService-Info | same | not committed |

CI may still produce unsigned or debug-signed artifacts. That is not store approval.
