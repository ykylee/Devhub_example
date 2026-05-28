/**
 * Vitest mock for react-dom/test-utils.
 * React 19 removes `act` from this module (it now throws).
 * @testing-library/react's act-compat.js falls back to this module;
 * we provide the real `act` from react's ESM export.
 */
import { act } from "react";
export { act };
const testUtils = { act };
export default testUtils;
