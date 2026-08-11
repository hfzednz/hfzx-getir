import { WarehouseDetailView } from "@/features/warehouses/components/warehouse-detail-view";

export default async function WarehouseDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <WarehouseDetailView warehouseId={id} />;
}
