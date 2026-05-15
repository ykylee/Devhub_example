import type { NextConfig } from "next";

import { BACKEND_API_URL_SERVER } from "./lib/config/endpoints";

// `output: "standalone"` 은 docker image minimize 용이지만 `next start` 와 호환되지
// 않아 e2e CI 의 webServer 와 native dev 가 깨진다. docker 빌드 시에만 활성화한다.
// 예: docker 의 build 단계에서 `NEXT_OUTPUT=standalone npm run build` 로 켠다.
const nextConfig: NextConfig = {
  output: process.env.NEXT_OUTPUT === "standalone" ? "standalone" : undefined,
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
