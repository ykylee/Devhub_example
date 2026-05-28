/**
 * Postinstall patches for React 19 × testing-library compatibility.
 *
 * React 19 removes `act` from the CJS `react` default export (named ESM only).
 * react-dom/test-utils.production.js wraps callback via `React.act(callback)`
 * which throws TypeError since React.act is undefined in CJS.
 *
 * Patch: add `act` to `react` CJS production module.exports so that
 * `React.act(callback)` works correctly for @testing-library/react.
 */
const fs = require("fs");
const path = require("path");

function appendToFile(filePath, content, marker, label) {
  if (!fs.existsSync(filePath)) {
    console.warn(`[postinstall] ${label} not found, skipping.`);
    return false;
  }
  const existing = fs.readFileSync(filePath, "utf8");
  if (existing.includes(marker)) {
    console.log(`[postinstall] ${label} already patched ✓`);
    return true;
  }
  fs.writeFileSync(filePath, existing + "\n" + content, "utf8");
  console.log(`[postinstall] Patched ${label} ✓`);
  return true;
}

const reactCjsProd = path.resolve(
  __dirname, "..",
  "node_modules/react/cjs/react.production.js",
);

appendToFile(
  reactCjsProd,
  [
    "// [postinstall] React.act polyfill for jsdom/vitest compatibility",
    "exports.act = function (callback) {",
    "  var result;",
    "  try {",
    '    var ReactDOM = require("react-dom");',
    "    ReactDOM.flushSync(function () {",
    "      result = callback();",
    "    });",
    "  } catch (e) {",
    "    result = callback();",
    "  }",
    "  if (result !== null && typeof result === 'object' && typeof result.then === 'function') {",
    "    return result;",
    "  }",
    '  return { then: function (resolve) { resolve(result); } };',
    "};",
  ].join("\n"),
  "// [postinstall] React.act polyfill",
  "react/cjs/react.production.js",
);
