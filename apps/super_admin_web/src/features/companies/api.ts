import { ALLOW_MOCK_FALLBACK } from "@/shared/config/platform";
import { apiClient, ApiError, platformPath } from "@/shared/api/client";
import type {
  CompaniesListResponse,
  CompanyDetail,
  CompanyListItem,
  CreateCompanyInput,
  UpdateCompanyInput,
} from "./types";

function delay(ms = 220): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

const MOCK_COMPANIES: CompanyListItem[] = [
  {
    id: "co_acme",
    legalName: "ACME Holdings A.Ş.",
    tradeName: "ACME",
    countryCode: "TR",
    status: "active",
    tenantCount: 3,
    primaryCurrency: "TRY",
    createdAt: "2023-05-10T09:00:00.000Z",
  },
  {
    id: "co_nova",
    legalName: "Nova Retail GmbH",
    tradeName: "Nova",
    countryCode: "DE",
    status: "active",
    tenantCount: 2,
    primaryCurrency: "EUR",
    createdAt: "2024-02-14T11:00:00.000Z",
  },
  {
    id: "co_orbit",
    legalName: "Orbit Logistics Ltd",
    tradeName: "Orbit",
    countryCode: "SG",
    status: "active",
    tenantCount: 1,
    primaryCurrency: "SGD",
    createdAt: "2023-09-01T08:00:00.000Z",
  },
  {
    id: "co_delta",
    legalName: "Delta Commerce Inc",
    tradeName: "Delta",
    countryCode: "US",
    status: "suspended",
    tenantCount: 1,
    primaryCurrency: "USD",
    createdAt: "2025-01-02T16:30:00.000Z",
  },
];

function mockDetail(id: string): CompanyDetail {
  const base =
    MOCK_COMPANIES.find((c) => c.id === id) ??
    ({
      ...MOCK_COMPANIES[0],
      id,
      legalName: `Company ${id}`,
      tradeName: id,
    } satisfies CompanyListItem);

  return {
    ...base,
    business: {
      industry: "Quick commerce",
      taxId: `${base.countryCode}-TAX-1001`,
      vatNumber: `VAT-${base.countryCode}-88991`,
      billingEmail: `billing@${base.tradeName.toLowerCase()}.example`,
      registeredAddress: `1 Platform Ave, ${base.countryCode}`,
    },
    tax: {
      defaultTaxEngine: base.countryCode === "TR" ? "gib" : "avalara",
      vatRegistered: true,
      withholdingEnabled: base.countryCode === "TR",
      fiscalYearStartMonth: 1,
    },
    locales: {
      defaultLocale: base.countryCode === "TR" ? "tr-TR" : "en-US",
      locales:
        base.countryCode === "TR"
          ? ["tr-TR", "en-US"]
          : base.countryCode === "DE"
            ? ["de-DE", "en-US"]
            : ["en-US"],
      timeZone:
        base.countryCode === "TR"
          ? "Europe/Istanbul"
          : base.countryCode === "DE"
            ? "Europe/Berlin"
            : base.countryCode === "SG"
              ? "Asia/Singapore"
              : "America/New_York",
      currencies: [base.primaryCurrency, "USD"].filter(
        (v, i, a) => a.indexOf(v) === i,
      ),
    },
    domains: [
      {
        id: `dom_${base.id}_1`,
        hostname: `${base.tradeName.toLowerCase()}.nexora.example`,
        verified: true,
        primary: true,
      },
      {
        id: `dom_${base.id}_2`,
        hostname: `admin.${base.tradeName.toLowerCase()}.example`,
        verified: base.status === "active",
        primary: false,
      },
    ],
    branding: {
      primaryColor: "#0B6E6E",
      secondaryColor: "#0f8585",
      logoUrl: `/assets/companies/${base.id}.svg`,
      faviconUrl: `/assets/companies/${base.id}-icon.svg`,
    },
  };
}

export async function fetchCompanies(): Promise<CompaniesListResponse> {
  try {
    return await apiClient<CompaniesListResponse>(platformPath("/companies"));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return {
        items: [...MOCK_COMPANIES],
        total: MOCK_COMPANIES.length,
        generatedAt: new Date().toISOString(),
      };
    }
    throw err;
  }
}

export async function fetchCompany(id: string): Promise<CompanyDetail> {
  try {
    return await apiClient<CompanyDetail>(platformPath(`/companies/${id}`));
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      return mockDetail(id);
    }
    throw err;
  }
}

export async function createCompany(
  input: CreateCompanyInput,
): Promise<CompanyListItem> {
  try {
    return await apiClient<CompanyListItem>(platformPath("/companies"), {
      method: "POST",
      body: input,
      idempotent: true,
    });
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const created: CompanyListItem = {
        id: `co_${input.tradeName.toLowerCase().replace(/\s+/g, "_")}`,
        legalName: input.legalName,
        tradeName: input.tradeName,
        countryCode: input.countryCode,
        status: "draft",
        tenantCount: 0,
        primaryCurrency: input.primaryCurrency,
        createdAt: new Date().toISOString(),
      };
      MOCK_COMPANIES.unshift(created);
      return created;
    }
    throw err;
  }
}

export async function updateCompany(
  id: string,
  input: UpdateCompanyInput,
): Promise<CompanyDetail> {
  try {
    return await apiClient<CompanyDetail>(platformPath(`/companies/${id}`), {
      method: "PATCH",
      body: input,
      idempotent: true,
    });
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = MOCK_COMPANIES.findIndex((c) => c.id === id);
      if (idx >= 0) {
        MOCK_COMPANIES[idx] = {
          ...MOCK_COMPANIES[idx],
          legalName: input.legalName ?? MOCK_COMPANIES[idx].legalName,
          tradeName: input.tradeName ?? MOCK_COMPANIES[idx].tradeName,
          status: input.status ?? MOCK_COMPANIES[idx].status,
        };
      }
      const detail = mockDetail(id);
      return {
        ...detail,
        business: { ...detail.business, ...input.business },
        tax: { ...detail.tax, ...input.tax },
        locales: { ...detail.locales, ...input.locales },
        branding: { ...detail.branding, ...input.branding },
      };
    }
    throw err;
  }
}

export async function deleteCompany(id: string): Promise<void> {
  try {
    await apiClient<void>(platformPath(`/companies/${id}`), {
      method: "DELETE",
      idempotent: true,
    });
  } catch (err) {
    if (!ALLOW_MOCK_FALLBACK) throw err;
    if (err instanceof ApiError || err instanceof TypeError) {
      await delay();
      const idx = MOCK_COMPANIES.findIndex((c) => c.id === id);
      if (idx >= 0) MOCK_COMPANIES.splice(idx, 1);
      return;
    }
    throw err;
  }
}
