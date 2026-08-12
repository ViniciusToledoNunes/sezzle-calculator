import type { Operation } from "../domain/operations";

type CalculateResponse = {
  operation: Operation;
  operands: number[];
  result: number;
};

type ErrorResponse = {
  error?: {
    code?: string;
    message?: string;
  };
};

const API_URL = import.meta.env.VITE_API_URL ?? "";
const REQUEST_TIMEOUT_MS = 8_000;

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly code = "request_failed",
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export async function calculate(operation: Operation, operands: number[]): Promise<number> {
  const controller = new AbortController();
  const timeout = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  try {
    const response = await fetch(`${API_URL}/api/v1/calculate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ operation, operands }),
      signal: controller.signal,
    });

    const body = (await response.json().catch(() => ({}))) as CalculateResponse & ErrorResponse;
    if (!response.ok) {
      throw new ApiError(
        body.error?.message ?? "The calculation could not be completed.",
        body.error?.code,
      );
    }
    if (typeof body.result !== "number" || !Number.isFinite(body.result)) {
      throw new ApiError("The server returned an invalid result.", "invalid_response");
    }
    return body.result;
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    if (error instanceof DOMException && error.name === "AbortError") {
      throw new ApiError("The request timed out. Please try again.", "timeout");
    }
    throw new ApiError("Unable to reach the calculator service. Please try again.", "network_error");
  } finally {
    window.clearTimeout(timeout);
  }
}
