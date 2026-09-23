/**
 * Shared mock factory functions for common module mocks used across test files.
 *
 * Note: Vitest's `vi.mock` is file-scoped and its factory runs before import
 * resolution, so these factories are invoked via `await import(...)` inside
 * async vi.mock factories (or via vi.hoisted in the test file).
 */

export function createI18nMock() {
  return {
    useTranslation: () => ({
      t: (key: string) => `translated:${key}`,
      i18n: { language: "en" },
    }),
  };
}

export function createSonnerMock() {
  return {
    toast: {
      success: vi.fn(),
      error: vi.fn(),
      loading: vi.fn(),
      promise: vi.fn(),
      dismiss: vi.fn(),
    },
  };
}

export function createErrorAlertMock() {
  return {
    ErrorAlert: () => null,
  };
}
