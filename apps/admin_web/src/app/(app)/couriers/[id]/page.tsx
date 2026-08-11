import { CourierDetailView } from "@/features/couriers/components/courier-detail-view";

export default async function CourierDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <CourierDetailView courierId={id} />;
}
