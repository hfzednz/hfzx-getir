import { TenantDetailView } from "@/features/tenants/components/tenant-detail-view";

export default async function TenantDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <TenantDetailView tenantId={id} />;
}
