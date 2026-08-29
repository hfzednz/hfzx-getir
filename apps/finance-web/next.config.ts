import type { NextConfig } from "next";

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const finance = (process.env.FINANCE_INTERNAL || "http://127.0.0.1:8091").replace(/\/$/, "");

const nextConfig: NextConfig = {
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/ledger/:path*", destination: `${finance}/v1/ledger/:path*` },
    ];
  },
};

export default nextConfig;
