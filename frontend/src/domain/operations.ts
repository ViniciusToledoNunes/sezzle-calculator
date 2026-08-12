export type Operation =
  | "add"
  | "subtract"
  | "multiply"
  | "divide"
  | "power"
  | "square_root"
  | "percentage";

export type OperationDefinition = {
  id: Operation;
  label: string;
  symbol: string;
  arity: 1 | 2;
  description: string;
};

export const OPERATIONS: readonly OperationDefinition[] = [
  { id: "add", label: "Add", symbol: "+", arity: 2, description: "Addition" },
  { id: "subtract", label: "Subtract", symbol: "−", arity: 2, description: "Subtraction" },
  { id: "multiply", label: "Multiply", symbol: "×", arity: 2, description: "Multiplication" },
  { id: "divide", label: "Divide", symbol: "÷", arity: 2, description: "Division" },
  { id: "power", label: "Power", symbol: "xʸ", arity: 2, description: "Exponentiation" },
  { id: "square_root", label: "Square root", symbol: "√", arity: 1, description: "Square root" },
  { id: "percentage", label: "Percentage", symbol: "%", arity: 1, description: "Convert to decimal" },
] as const;

export function getOperation(id: Operation): OperationDefinition {
  const operation = OPERATIONS.find((candidate) => candidate.id === id);
  if (!operation) {
    throw new Error(`Unsupported operation: ${id}`);
  }
  return operation;
}

export function formatExpression(operation: OperationDefinition, operands: number[]): string {
  if (operation.id === "square_root") {
    return `√${formatNumber(operands[0])}`;
  }
  if (operation.id === "percentage") {
    return `${formatNumber(operands[0])}%`;
  }
  return `${formatNumber(operands[0])} ${operation.symbol} ${formatNumber(operands[1])}`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-US", {
    maximumSignificantDigits: 12,
    useGrouping: true,
  }).format(value);
}
