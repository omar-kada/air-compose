import {
  formatElapsed,
  formatElapsedMs,
  humanizeDuration,
  humanizeDurationMs,
  isDateValid,
} from "./time";

describe("isDateValid", () => {
  it("returns false for undefined", () => {
    expect(isDateValid(undefined)).toBe(false);
  });

  it("returns false for unparseable strings", () => {
    expect(isDateValid("not-a-date")).toBe(false);
  });

  it("returns true for valid ISO strings", () => {
    expect(isDateValid("2024-01-15T10:30:45.000Z")).toBe(true);
  });

  it("returns true for valid Date instances", () => {
    expect(isDateValid(new Date("2024-01-15"))).toBe(true);
  });
});

describe("formatElapsed", () => {
  it('returns "-" when either date is invalid', () => {
    expect(formatElapsed(undefined, undefined)).toBe("-");
    expect(formatElapsed("not-a-date", "2024-01-15T00:00:00Z")).toBe("-");
    expect(formatElapsed("2024-01-15T00:00:00Z", undefined)).toBe("-");
  });

  it('formats a 5-second span as "5s"', () => {
    expect(formatElapsed("2024-01-15T00:00:00Z", "2024-01-15T00:00:05Z")).toBe(
      "5s",
    );
  });

  it('formats a 2-minute span as "2m"', () => {
    expect(formatElapsed("2024-01-15T00:00:00Z", "2024-01-15T00:02:00Z")).toBe(
      "2m",
    );
  });
});

describe("formatElapsedMs", () => {
  it('formats zero as "0s"', () => {
    expect(formatElapsedMs(0)).toBe("0s");
  });

  it("formats seconds", () => {
    expect(formatElapsedMs(5000)).toBe("5s");
  });

  it("rounds to the nearest minute", () => {
    expect(formatElapsedMs(120000)).toBe("2m");
    expect(formatElapsedMs(90000)).toBe("2m"); // 1.5 minutes -> rounds to 2
  });

  it("formats hours", () => {
    expect(formatElapsedMs(3600000)).toBe("1h");
  });
});

describe("humanizeDurationMs", () => {
  it('returns "now" for zero', () => {
    expect(humanizeDurationMs(0)).toBe("now");
  });

  it('formats future seconds as "in N seconds"', () => {
    expect(humanizeDurationMs(5000, "en")).toBe("in 5 seconds");
  });

  it('formats past seconds as "N seconds ago"', () => {
    expect(humanizeDurationMs(-5000, "en")).toBe("5 seconds ago");
  });

  it("formats minutes", () => {
    expect(humanizeDurationMs(120000, "en")).toBe("in 2 minutes");
  });
});

describe("humanizeDuration", () => {
  it("computes the relative duration between two dates", () => {
    expect(
      humanizeDuration("2024-01-15T00:00:00Z", "2024-01-15T00:00:05Z"),
    ).toBe("in 5 seconds");
  });
});
