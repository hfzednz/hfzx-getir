import type { NextConfig } from "next";

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const supplier = (process.env.SUPPLIER_INTERNAL || "http://127.0.0.1:8117").replace(/\/$/, "");

const nextConfig: NextConfig = {
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/supplier/:path*", destination: `${supplier}/v1/supplier/:path*` },
    ];
  },
};

export default nextConfig;
