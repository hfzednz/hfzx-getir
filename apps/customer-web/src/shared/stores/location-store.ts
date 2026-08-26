"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

interface LocationState {
  lat: number;
  lng: number;
  addressLabel: string;
  setLocation: (lat: number, lng: number, label: string) => void;
}

export const useLocation = create<LocationState>()(
  persist(
    (set) => ({
      lat: 41.0082,
      lng: 28.9784,
      addressLabel: "Istanbul",
      setLocation: (lat, lng, addressLabel) => set({ lat, lng, addressLabel }),
    }),
    { name: "nexora-customer-location" },
  ),
);
