import type { Id } from "@/shared/types/common";

export type CompanyStatus = "active" | "suspended" | "draft";

export interface CompanyListItem {
  id: Id;
  legalName: string;
  tradeName: string;
  countryCode: string;
  status: CompanyStatus;
  tenantCount: number;
  primaryCurrency: string;
  createdAt: string;
}

export interface CompanyBusinessSettings {
  industry: string;
  taxId: string;
  vatNumber: string;
  billingEmail: string;
  registeredAddress: string;
}

export interface CompanyTaxSettings {
  defaultTaxEngine: string;
  vatRegistered: boolean;
  withholdingEnabled: boolean;
  fiscalYearStartMonth: number;
}

export interface CompanyLocaleSettings {
  defaultLocale: string;
  locales: string[];
  timeZone: string;
  currencies: string[];
}

export interface CompanyDomain {
  id: Id;
  hostname: string;
  verified: boolean;
  primary: boolean;
}

export interface CompanyBranding {
  primaryColor: string;
  secondaryColor: string;
  logoUrl: string;
  faviconUrl: string;
}

export interface CompanyDetail extends CompanyListItem {
  business: CompanyBusinessSettings;
  tax: CompanyTaxSettings;
  locales: CompanyLocaleSettings;
  domains: CompanyDomain[];
  branding: CompanyBranding;
}

export interface CreateCompanyInput {
  legalName: string;
  tradeName: string;
  countryCode: string;
  primaryCurrency: string;
}

export interface UpdateCompanyInput {
  legalName?: string;
  tradeName?: string;
  status?: CompanyStatus;
  business?: Partial<CompanyBusinessSettings>;
  tax?: Partial<CompanyTaxSettings>;
  locales?: Partial<CompanyLocaleSettings>;
  branding?: Partial<CompanyBranding>;
}

export interface CompaniesListResponse {
  items: CompanyListItem[];
  total: number;
  generatedAt: string;
}
