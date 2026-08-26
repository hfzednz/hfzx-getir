"use client";
import { useRouter } from "next/navigation";
import { useSession } from "@/shared/api/client";
export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);
  return (
    <div className="flex min-h-dvh flex-col justify-center p-6">
      <h1 className="mb-6 text-2xl font-bold">Courier</h1>
      <button type="button" className="rounded-lg bg-violet-600 py-3 font-semibold"
        onClick={() => { setSession({ accessToken: "demo", principalId: "courier-1", roles: ["courier"] }); router.push("/dashboard"); }}>
        Sign in (courier)
      </button>
    </div>
  );
}
