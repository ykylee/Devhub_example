import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

// Vitest config (PR-T2, work_26_05_11-d sprint).
//   - jsdom env so jest-dom matchers + DOM globals work the same as in
//     Next.js client components
//   - alias mirrors the `@/*` path mapping from tsconfig.json
//   - globals = true so vi/describe/it/expect/beforeEach are available
//     without per-file imports (matches the M1-spec style for Go tests)
export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./lib/test-setup.ts"],
    include: ["lib/**/*.test.ts", "lib/**/*.test.tsx", "components/**/*.test.tsx"],
    exclude: ["node_modules", ".next", "playwright-report", "tests/e2e"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      include: ["lib/**/*.ts", "components/**/*.tsx"],
      exclude: ["**/*.test.ts", "**/*.test.tsx", "**/test-setup.ts"],
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "."),
    },
  },
});
