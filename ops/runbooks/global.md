# Global / localization runbook

## Activate a new country

1. `POST /v1/global/countries` then `.../activate`
2. Seed locales, currencies, tax, privacy, payment availability, rules
3. Verify `GET /v1/global/resolve?country=XX&locale=...`

## FX refresh

`POST /v1/global/rates/refresh` — display FX only; settlement rates remain in finance/payment domains.
