import type { NextConfig } from "next";

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const platform = (process.env.PLATFORM_OPS_INTERNAL || "http://127.0.0.1:8110").replace(/\/$/, "");

const nextConfig: NextConfig = {
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/platform/:path*", destination: `${platform}/v1/platform/:path*` },
    ];
  },
};

export default nextConfig;
