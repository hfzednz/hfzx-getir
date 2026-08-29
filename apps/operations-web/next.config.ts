import path from "node:path";
import type { NextConfig } from "next";

// The shared @nexora/ui and @nexora/web-core packages live outside this app, so
// Turbopack has to treat the monorepo root as its project root.
const workspaceRoot = path.resolve(__dirname, "../..");

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const bff = (process.env.BFF_ADMIN_INTERNAL || "http://127.0.0.1:8114").replace(/\/$/, "");

const nextConfig: NextConfig = {
  turbopack: { root: workspaceRoot },
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/:path*", destination: `${bff}/v1/:path*` },
    ];
  },
};

export default nextConfig;
