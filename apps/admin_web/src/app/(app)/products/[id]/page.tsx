import { ProductDetailView } from "@/features/products/components/product-detail-view";

export default async function ProductDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <ProductDetailView productId={id} />;
}
