import { CountryDetailView } from "@/features/countries/components/country-detail-view";

export default async function CountryDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <CountryDetailView countryId={id} />;
}
