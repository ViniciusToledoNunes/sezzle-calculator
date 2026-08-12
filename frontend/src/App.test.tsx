import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";
import { ApiError, calculate } from "./api/calculator";

vi.mock("./api/calculator", async (importOriginal) => {
  const original = await importOriginal<typeof import("./api/calculator")>();
  return { ...original, calculate: vi.fn() };
});

const mockedCalculate = vi.mocked(calculate);

describe("App", () => {
  beforeEach(() => mockedCalculate.mockReset());

  it("validates required values before calling the API", async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: "Calculate result" }));

    expect(screen.getByText("Enter the first number.")).toBeInTheDocument();
    expect(screen.getByText("Enter the second number.")).toBeInTheDocument();
    expect(mockedCalculate).not.toHaveBeenCalled();
  });

  it("calculates a binary operation and adds it to history", async () => {
    mockedCalculate.mockResolvedValue(42);
    const user = userEvent.setup();
    render(<App />);

    await user.type(screen.getByLabelText("First number"), "6");
    await user.type(screen.getByLabelText("Second number"), "7");
    await user.click(screen.getByRole("button", { name: "Multiply" }));
    await user.click(screen.getByRole("button", { name: "Calculate result" }));

    await waitFor(() => expect(mockedCalculate).toHaveBeenCalledWith("multiply", [6, 7]));
    expect(await screen.findByLabelText("Result: 42")).toBeInTheDocument();
    expect(screen.getByText("6 × 7")).toBeInTheDocument();
  });

  it("sends one operand for unary operations", async () => {
    mockedCalculate.mockResolvedValue(9);
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole("button", { name: "Square root" }));
    await user.type(screen.getByLabelText("Number"), "81");
    expect(screen.queryByLabelText("Second number")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Calculate result" }));

    await waitFor(() => expect(mockedCalculate).toHaveBeenCalledWith("square_root", [81]));
    expect(await screen.findByLabelText("Result: 9")).toBeInTheDocument();
  });

  it("shows service errors and allows clearing the form", async () => {
    mockedCalculate.mockRejectedValueOnce(new ApiError("Division by zero is not allowed.", "division_by_zero"));
    const user = userEvent.setup();
    render(<App />);

    await user.type(screen.getByLabelText("First number"), "3");
    await user.type(screen.getByLabelText("Second number"), "0");
    await user.click(screen.getByRole("button", { name: "Divide" }));
    await user.click(screen.getByRole("button", { name: "Calculate result" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Division by zero is not allowed.");
    await user.click(screen.getAllByRole("button", { name: "Clear" })[0]);
    expect(screen.getByLabelText("First number")).toHaveValue("");
  });
});
