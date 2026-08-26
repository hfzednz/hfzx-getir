"use client";
import { useRouter } from "next/navigation";
import { useSession } from "@/shared/api/client";
export default function LoginPage() {
  const router = useRouter();
  const setSession = useSession((s) => s.setSession);
  return (
    <div className="flex min-h-dvh flex-col justify-center p-6 max-w-md mx-auto">
      <h1 className="mb-6 text-2xl font-bold text-violet-700">Warehouse</h1>
      <button type="button" className="rounded-lg bg-violet-600 py-3 font-semibold text-white"
        onClick={() => { setSession({ accessToken: "demo", principalId: "picker-1", roles: ["picker"] }); router.push("/dashboard"); }}>
        Sign in as picker
      </button>
    </div>
  );
}
