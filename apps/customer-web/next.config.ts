import type { NextConfig } from "next";

const bffInternal = (process.env.BFF_CUSTOMER_INTERNAL || "http://127.0.0.1:8111").replace(
  /\/$/,
  "",
);
const realtimeInternal = (process.env.REALTIME_INTERNAL || "http://127.0.0.1:8115").replace(
  /\/$/,
  "",
);

const nextConfig: NextConfig = {
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/realtime/:path*", destination: `${realtimeInternal}/v1/realtime/:path*` },
      { source: "/v1/:path*", destination: `${bffInternal}/v1/:path*` },
      { source: "/health", destination: `${bffInternal}/health` },
    ];
  },
};

export default nextConfig;
