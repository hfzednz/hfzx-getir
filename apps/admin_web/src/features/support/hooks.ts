"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  approveTicketRefund,
  escalateTicket,
  fetchSupportWorkspace,
  fetchTicket,
  resolveTicket,
} from "./api";
import type { TicketListParams } from "./types";

export const supportKeys = {
  all: ["support"] as const,
  workspace: (params: TicketListParams) =>
    [...supportKeys.all, "workspace", params] as const,
  ticket: (id: string) => [...supportKeys.all, "ticket", id] as const,
};

export function useSupportWorkspace(params: TicketListParams = {}) {
  return useQuery({
    queryKey: supportKeys.workspace(params),
    queryFn: () => fetchSupportWorkspace(params),
  });
}

export function useTicket(id: string) {
  return useQuery({
    queryKey: supportKeys.ticket(id),
    queryFn: () => fetchTicket(id),
    enabled: Boolean(id),
  });
}

function invalidateSupport(
  qc: ReturnType<typeof useQueryClient>,
  id?: string,
) {
  void qc.invalidateQueries({ queryKey: supportKeys.all });
  if (id) void qc.invalidateQueries({ queryKey: supportKeys.ticket(id) });
}

export function useEscalateTicket(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => escalateTicket(id),
    onSuccess: () => invalidateSupport(qc, id),
  });
}

export function useResolveTicket(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => resolveTicket(id),
    onSuccess: () => invalidateSupport(qc, id),
  });
}

export function useApproveTicketRefund(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => approveTicketRefund(id),
    onSuccess: () => invalidateSupport(qc, id),
  });
}
