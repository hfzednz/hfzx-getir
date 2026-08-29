import path from "node:path";
import type { NextConfig } from "next";

// The shared @nexora/ui and @nexora/web-core packages live outside this app, so
// Turbopack has to treat the monorepo root as its project root.
const workspaceRoot = path.resolve(__dirname, "../..");

const identity = (process.env.IDENTITY_INTERNAL || "http://127.0.0.1:8081").replace(/\/$/, "");
const supplier = (process.env.SUPPLIER_INTERNAL || "http://127.0.0.1:8117").replace(/\/$/, "");

const nextConfig: NextConfig = {
  turbopack: { root: workspaceRoot },
  transpilePackages: ["@nexora/ui", "@nexora/web-core"],
  async rewrites() {
    return [
      { source: "/v1/identity/:path*", destination: `${identity}/v1/identity/:path*` },
      { source: "/v1/supplier/:path*", destination: `${supplier}/v1/supplier/:path*` },
    ];
  },
};

export default nextConfig;
