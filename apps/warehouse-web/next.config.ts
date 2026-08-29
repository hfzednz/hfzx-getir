import type { NextConfig } from "next";

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const bff = (process.env.BFF_WAREHOUSE_INTERNAL || "http://127.0.0.1:8113").replace(/\/$/, "");

const nextConfig: NextConfig = {
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/:path*", destination: `${bff}/v1/:path*` },
    ];
  },
};

export default nextConfig;
