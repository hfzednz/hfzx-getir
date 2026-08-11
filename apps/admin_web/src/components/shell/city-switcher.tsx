"use client";

import { useMemo } from "react";
import type { City } from "@/shared/types/common";
import { Select } from "@nexora/ui";
import { useChromeStore } from "@/stores/chrome-store";

const DEMO_CITIES: City[] = [
  { id: "city_ist", name: "Istanbul", code: "IST", active: true },
  { id: "city_ank", name: "Ankara", code: "ANK", active: true },
  { id: "city_izm", name: "Izmir", code: "IZM", active: true },
  { id: "city_all", name: "All cities", code: "ALL", active: true },
];

export function CitySwitcher({ cities = DEMO_CITIES }: { cities?: City[] }) {
  const cityId = useChromeStore((s) => s.cityId);
  const setCityId = useChromeStore((s) => s.setCityId);

  const value = useMemo(() => {
    if (cityId && cities.some((c) => c.id === cityId)) return cityId;
    return cities[0]?.id ?? "";
  }, [cityId, cities]);

  return (
    <label className="flex items-center gap-[var(--nx-space-2)]">
      <span className="sr-only">City</span>
      <Select
        aria-label="City switcher"
        value={value}
        onChange={(e) => setCityId(e.target.value || null)}
        className="min-w-[140px] h-[var(--nx-control-height-sm)] text-[12px]"
      >
        {cities.map((city) => (
          <option key={city.id} value={city.id}>
            {city.name}
          </option>
        ))}
      </Select>
    </label>
  );
}
