import { CompanyDetailView } from "@/features/companies/components/company-detail-view";

export default async function CompanyDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <CompanyDetailView companyId={id} />;
}
