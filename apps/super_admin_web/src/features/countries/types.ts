import type { Id } from "@/shared/types/common";

export interface CountryListItem {
  id: Id;
  code: string;
  name: string;
  regionCount: number;
  cityCount: number;
  defaultCurrency: string;
  defaultLocale: string;
  status: "active" | "pilot" | "disabled";
}

export interface CountryLanguage {
  code: string;
  label: string;
  primary: boolean;
}

export interface CountryCurrency {
  code: string;
  label: string;
  primary: boolean;
}

export interface CountryTimezone {
  id: string;
  label: string;
  offset: string;
}

export interface CountryTaxRule {
  id: Id;
  name: string;
  ratePct: number;
  appliesTo: string;
}

export interface DeliveryRulesSummary {
  maxRadiusKm: number;
  defaultSlaMin: number;
  nightSurchargePct: number;
  zoneEditorNote: string;
}

export interface LegalRule {
  id: Id;
  framework: string;
  summary: string;
  status: "active" | "draft";
}

export interface Holiday {
  id: Id;
  name: string;
  date: string;
  affectsDelivery: boolean;
}

export interface RegionalPricingHook {
  id: Id;
  key: string;
  description: string;
  enabled: boolean;
}

export interface CountryCity {
  id: Id;
  name: string;
  warehouseCount: number;
  status: "active" | "planned";
}

export interface CountryRegion {
  id: Id;
  name: string;
  cities: CountryCity[];
}

export interface CountryDetail extends CountryListItem {
  languages: CountryLanguage[];
  currencies: CountryCurrency[];
  timezones: CountryTimezone[];
  taxes: CountryTaxRule[];
  deliveryRules: DeliveryRulesSummary;
  legalRules: LegalRule[];
  holidays: Holiday[];
  regionalPricingHooks: RegionalPricingHook[];
  regions: CountryRegion[];
}

export interface CountriesListResponse {
  items: CountryListItem[];
  total: number;
  generatedAt: string;
}
