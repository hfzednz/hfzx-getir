import { TicketDetailView } from "@/features/support/components/ticket-detail-view";

export default async function SupportTicketPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <TicketDetailView id={id} />;
}
