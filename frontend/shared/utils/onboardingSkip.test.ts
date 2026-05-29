import { afterEach, describe, expect, it } from "vitest";
import {
  isOnboardingSkipped,
  markOnboardingSkipped,
  clearOnboardingSkip,
} from "./onboardingSkip";

const KEY = "devhub.onboarding.skipped";

describe("onboardingSkip", () => {
  afterEach(() => {
    if (typeof window !== "undefined") {
      try {
        window.sessionStorage.removeItem(KEY);
      } catch {
        /* ignore */
      }
    }
  });

  describe("isOnboardingSkipped", () => {
    it("returns false when sessionStorage value is not set", () => {
      expect(isOnboardingSkipped()).toBe(false);
    });

    it("returns true when sessionStorage value is '1'", () => {
      window.sessionStorage.setItem(KEY, "1");
      expect(isOnboardingSkipped()).toBe(true);
    });

    it("returns false when sessionStorage value is anything other than '1'", () => {
      window.sessionStorage.setItem(KEY, "true");
      expect(isOnboardingSkipped()).toBe(false);
    });
  });

  describe("markOnboardingSkipped", () => {
    it("sets sessionStorage value to '1'", () => {
      markOnboardingSkipped();
      expect(window.sessionStorage.getItem(KEY)).toBe("1");
      expect(isOnboardingSkipped()).toBe(true);
    });

    it("is idempotent on repeated calls", () => {
      markOnboardingSkipped();
      markOnboardingSkipped();
      expect(window.sessionStorage.getItem(KEY)).toBe("1");
    });
  });

  describe("clearOnboardingSkip", () => {
    it("removes the sessionStorage value", () => {
      window.sessionStorage.setItem(KEY, "1");
      clearOnboardingSkip();
      expect(window.sessionStorage.getItem(KEY)).toBeNull();
      expect(isOnboardingSkipped()).toBe(false);
    });

    it("is a no-op when no value has been set", () => {
      clearOnboardingSkip();
      expect(window.sessionStorage.getItem(KEY)).toBeNull();
    });
  });

  describe("interaction sequence", () => {
    it("mark → check → clear → check cycle", () => {
      expect(isOnboardingSkipped()).toBe(false);
      markOnboardingSkipped();
      expect(isOnboardingSkipped()).toBe(true);
      clearOnboardingSkip();
      expect(isOnboardingSkipped()).toBe(false);
    });
  });

  describe("SSR environment fallback", () => {
    it("gracefully falls back when window is undefined", () => {
      vi.stubGlobal("window", undefined);
      
      expect(isOnboardingSkipped()).toBe(false);
      expect(() => markOnboardingSkipped()).not.toThrow();
      expect(() => clearOnboardingSkip()).not.toThrow();
      
      vi.unstubAllGlobals();
    });
  });

  describe("sessionStorage error boundaries", () => {
    it("swallows errors thrown by sessionStorage operations", () => {
      const originalStorage = window.sessionStorage;
      
      // Override sessionStorage property to force runtime throws in get/set/remove operations
      Object.defineProperty(window, "sessionStorage", {
        get() {
          throw new Error("forced sandboxed error");
        },
        configurable: true,
      });

      expect(isOnboardingSkipped()).toBe(false);
      expect(() => markOnboardingSkipped()).not.toThrow();
      expect(() => clearOnboardingSkip()).not.toThrow();

      // Restore original sessionStorage
      Object.defineProperty(window, "sessionStorage", {
        value: originalStorage,
        configurable: true,
      });
    });
  });
});
