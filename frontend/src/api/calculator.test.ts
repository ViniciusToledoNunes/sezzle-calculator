import { afterEach, describe, expect, it, vi } from "vitest";
import { calculate } from "./calculator";

describe("calculator API client", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("returns a valid result", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ operation: "add", operands: [2, 3], result: 5 }),
    });
    vi.stubGlobal("fetch", fetchMock);

    await expect(calculate("add", [2, 3])).resolves.toBe(5);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/calculate", expect.objectContaining({ method: "POST" }));
  });

  it("preserves API error details", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: { code: "division_by_zero", message: "Division by zero." } }),
    }));

    await expect(calculate("divide", [1, 0])).rejects.toEqual(
      expect.objectContaining({ code: "division_by_zero", message: "Division by zero." }),
    );
  });

  it("rejects malformed successful responses", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ result: "five" }),
    }));

    await expect(calculate("add", [2, 3])).rejects.toMatchObject({ code: "invalid_response" });
  });

  it("turns network failures into a useful error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("offline")));
    await expect(calculate("add", [2, 3])).rejects.toMatchObject({ code: "network_error" });
  });
});
