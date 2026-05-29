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
    environment: "happy-dom",
    globals: true,
    setupFiles: ["./lib/test-setup.ts"],
    include: ["lib/**/*.test.ts", "lib/**/*.test.tsx", "components/**/*.test.tsx", "domain/**/*.test.ts", "domain/**/*.test.tsx", "shared/**/*.test.ts", "shared/**/*.test.tsx"],
    exclude: ["node_modules", ".next", "playwright-report", "tests/e2e"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      include: ["lib/**/*.ts", "components/**/*.tsx", "domain/**/*.ts", "domain/**/*.tsx", "shared/**/*.ts", "shared/**/*.tsx"],
      exclude: ["**/*.test.ts", "**/*.test.tsx", "**/test-setup.ts", "lib/__mocks__/**"],
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "."),
      // React 19 removes act from react-dom/test-utils. Redirect to our mock
      // so @testing-library/react's act-compat.js gets the real act from react.
      "react-dom/test-utils": path.resolve(__dirname, "lib/__mocks__/react-dom-test-utils.ts"),
    },
  },
});
