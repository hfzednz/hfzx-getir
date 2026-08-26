import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  CountriesListResponse,
  CountryDetail,
  CountryListItem,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_COUNTRIES: CountryListItem[] = [
  {
    id: "cty_tr",
    code: "TR",
    name: "Türkiye",
    regionCount: 7,
    cityCount: 12,
    defaultCurrency: "TRY",
    defaultLocale: "tr-TR",
    status: "active",
  },
  {
    id: "cty_de",
    code: "DE",
    name: "Germany",
    regionCount: 4,
    cityCount: 8,
    defaultCurrency: "EUR",
    defaultLocale: "de-DE",
    status: "active",
  },
  {
    id: "cty_sg",
    code: "SG",
    name: "Singapore",
    regionCount: 1,
    cityCount: 1,
    defaultCurrency: "SGD",
    defaultLocale: "en-SG",
    status: "pilot",
  },
  {
    id: "cty_us",
    code: "US",
    name: "United States",
    regionCount: 3,
    cityCount: 6,
    defaultCurrency: "USD",
    defaultLocale: "en-US",
    status: "disabled",
  },
];

function mockDetail(id: string): CountryDetail {
  const base =
    MOCK_COUNTRIES.find((c) => c.id === id) ??
    ({
      ...MOCK_COUNTRIES[0],
      id,
      name: `Country ${id}`,
    } satisfies CountryListItem);

  return {
    ...base,
    languages: [
      { code: base.defaultLocale, label: "Primary", primary: true },
      { code: "en-US", label: "English", primary: false },
    ],
    currencies: [
      { code: base.defaultCurrency, label: "Primary", primary: true },
      { code: "USD", label: "USD settlement", primary: false },
    ],
    timezones: [
      {
        id: "tz1",
        label:
          base.code === "TR"
            ? "Europe/Istanbul"
            : base.code === "DE"
              ? "Europe/Berlin"
              : base.code === "SG"
                ? "Asia/Singapore"
                : "America/New_York",
        offset: base.code === "US" ? "UTC-5" : "UTC+3",
      },
    ],
    taxes: [
      {
        id: "tax1",
        name: "Standard VAT/GST",
        ratePct: base.code === "TR" ? 20 : base.code === "DE" ? 19 : 8,
        appliesTo: "Most goods",
      },
      {
        id: "tax2",
        name: "Reduced rate",
        ratePct: base.code === "TR" ? 10 : 7,
        appliesTo: "Food staples",
      },
    ],
    deliveryRules: {
      maxRadiusKm: 8,
      defaultSlaMin: 15,
      nightSurchargePct: 15,
      zoneEditorNote:
        "Delivery zone polygons are edited in Admin Web city-ops — Super Admin shows summary only.",
    },
    legalRules: [
      {
        id: "leg1",
        framework: base.code === "TR" ? "KVKK" : "GDPR/CCPA",
        summary: "Platform privacy baseline for this country",
        status: "active",
      },
      {
        id: "leg2",
        framework: "Consumer protection",
        summary: "Cooling-off and refund disclosure requirements",
        status: "active",
      },
    ],
    holidays: [
      {
        id: "hol1",
        name: "National day",
        date: "2026-10-29",
        affectsDelivery: true,
      },
      {
        id: "hol2",
        name: "New Year",
        date: "2027-01-01",
        affectsDelivery: true,
      },
    ],
    regionalPricingHooks: [
      {
        id: "rp1",
        key: "surge.weather",
        description: "Weather-based regional pricing multiplier",
        enabled: true,
      },
      {
        id: "rp2",
        key: "fx.settle",
        description: "FX settlement floor for multi-currency carts",
        enabled: base.code !== "US",
      },
    ],
    regions: [
      {
        id: "reg1",
        name: base.code === "TR" ? "Marmara" : "Primary region",
        cities: [
          {
            id: "city1",
            name: base.code === "TR" ? "İstanbul" : base.name + " City",
            warehouseCount: 14,
            status: "active",
          },
          {
            id: "city2",
            name: base.code === "TR" ? "Bursa" : "Secondary City",
            warehouseCount: 3,
            status: "active",
          },
        ],
      },
      {
        id: "reg2",
        name: base.code === "TR" ? "İç Anadolu" : "Expansion",
        cities: [
          {
            id: "city3",
            name: base.code === "TR" ? "Ankara" : "Pilot City",
            warehouseCount: 6,
            status: base.status === "disabled" ? "planned" : "active",
          },
        ],
      },
    ],
  };
}

export async function fetchCountries(): Promise<CountriesListResponse> {
  try {
    return await apiClient<CountriesListResponse>(platformPath("/countries"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return {
        items: [...MOCK_COUNTRIES],
        total: MOCK_COUNTRIES.length,
        generatedAt: new Date().toISOString(),
      };
    }
    throw err;
  }
}

export async function fetchCountry(id: string): Promise<CountryDetail> {
  try {
    return await apiClient<CountryDetail>(platformPath(`/countries/${id}`));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockDetail(id);
    }
    throw err;
  }
}
