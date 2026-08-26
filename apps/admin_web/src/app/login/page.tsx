"use client";

import { useRouter } from "next/navigation";
import { OtpLoginForm } from "@nexora/web-core";
import { useAuthStore } from "@/shared/auth/auth-store";

export default function LoginPage() {
  const router = useRouter();
  const setSessionFromOtp = useAuthStore((s) => s.setSessionFromOtp);

  return (
    <div className="min-h-screen flex items-center justify-center px-[var(--nx-space-4)] bg-[var(--nx-bg-canvas)]">
      <div className="w-full max-w-sm">
        <OtpLoginForm
          title="NEXORA Admin"
          expectedRoles={["admin", "city_ops", "support_agent", "finance_analyst"]}
          onSuccess={(session) => {
            setSessionFromOtp(session, session.phone ?? "+905551112233");
            router.replace("/dashboard");
          }}
          otpHint="Staging: OTP from identity-service logs when OTP_DEV_MODE=true."
        />
      </div>
    </div>
  );
}
