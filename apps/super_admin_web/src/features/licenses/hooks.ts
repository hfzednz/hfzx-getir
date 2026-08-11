"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchLicenses,
  proposeLicenseOverride,
  renewLicense,
  resolveLicenseProposal,
} from "./api";

export const licenseKeys = {
  all: ["platform-licenses"] as const,
  snapshot: () => [...licenseKeys.all, "snapshot"] as const,
};

export function useLicenses() {
  return useQuery({
    queryKey: licenseKeys.snapshot(),
    queryFn: fetchLicenses,
  });
}

export function useRenewLicense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (licenseId: string) => renewLicense(licenseId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: licenseKeys.all });
    },
  });
}

export function useProposeLicenseOverride() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: proposeLicenseOverride,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: licenseKeys.all });
    },
  });
}

export function useResolveLicenseProposal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: resolveLicenseProposal,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: licenseKeys.all });
    },
  });
}
