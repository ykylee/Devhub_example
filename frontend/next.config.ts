import type { NextConfig } from "next";

const BACKEND_URL = process.env.BACKEND_API_URL || "http://backend-core:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND_URL}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
