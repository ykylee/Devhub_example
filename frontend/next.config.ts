import type { NextConfig } from "next";

import { BACKEND_API_URL_SERVER } from "@/shared/config/endpoints";

// `output: "standalone"` 은 host build 후 runtime-only Docker image 를 만들 때 사용한다.
// native dev 와는 분리되므로, 배포 패키징 스크립트가 host build 단계에서만 켠다.
const nextConfig: NextConfig = {
  basePath: process.env.NEXT_PUBLIC_BASE_PATH ? `/${process.env.NEXT_PUBLIC_BASE_PATH.replace(/^\//, "")}` : undefined,
  output: process.env.NEXT_OUTPUT === "standalone" ? "standalone" : undefined,
  turbopack: {
    root: process.cwd(),
  },
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
