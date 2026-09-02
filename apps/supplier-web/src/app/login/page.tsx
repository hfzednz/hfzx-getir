"use client";
import { useRouter } from "next/navigation";
import { OtpLoginForm } from "@nexora/web-core";
import { useSession } from "@/shared/api/client";

export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);
  return (
    <OtpLoginForm
      title="Supplier / Merchant"
      expectedRoles={["supplier", "partner", "merchant"]}
      onSuccess={(s) => { setSession(s); router.push("/dashboard"); }}
    />
  );
}
