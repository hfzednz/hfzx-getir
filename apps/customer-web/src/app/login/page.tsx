"use client";

import { useRouter } from "next/navigation";
import { OtpLoginForm } from "@nexora/web-core";
import { useSession } from "@/shared/api/client";

export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);

  return (
    <OtpLoginForm
      title="NEXORA"
      channel="customer-bff"
      expectedRoles={["customer"]}
      onSuccess={(session) => {
        setSession(session);
        router.push("/home");
      }}
    />
  );
}
