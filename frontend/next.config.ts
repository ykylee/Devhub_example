import type { NextConfig } from "next";

import { BACKEND_API_URL_SERVER } from "./lib/config/endpoints";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND_API_URL_SERVER}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
