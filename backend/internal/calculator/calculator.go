package calculator

import (
	"errors"
	"math"
)

type Operation string

const (
	Add        Operation = "add"
	Subtract   Operation = "subtract"
	Multiply   Operation = "multiply"
	Divide     Operation = "divide"
	Power      Operation = "power"
	SquareRoot Operation = "square_root"
	Percentage Operation = "percentage"
)

var (
	ErrUnknownOperation    = errors.New("unknown operation")
	ErrInvalidOperandCount = errors.New("invalid operand count")
	ErrInvalidOperand      = errors.New("operands must be finite numbers")
	ErrDivisionByZero      = errors.New("division by zero is not allowed")
	ErrNegativeSquareRoot  = errors.New("square root is only defined for non-negative numbers")
	ErrUndefinedResult     = errors.New("the calculation does not produce a finite result")
)

// Calculate evaluates an operation with exactly the number of operands it needs.
func Calculate(operation Operation, operands []float64) (float64, error) {
	want, ok := operandCount(operation)
	if !ok {
		return 0, ErrUnknownOperation
	}
	if len(operands) != want {
		return 0, ErrInvalidOperandCount
	}
	for _, operand := range operands {
		if math.IsNaN(operand) || math.IsInf(operand, 0) {
			return 0, ErrInvalidOperand
		}
	}

	var result float64
	switch operation {
	case Add:
		result = operands[0] + operands[1]
	case Subtract:
		result = operands[0] - operands[1]
	case Multiply:
		result = operands[0] * operands[1]
	case Divide:
		if operands[1] == 0 {
			return 0, ErrDivisionByZero
		}
		result = operands[0] / operands[1]
	case Power:
		result = math.Pow(operands[0], operands[1])
	case SquareRoot:
		if operands[0] < 0 {
			return 0, ErrNegativeSquareRoot
		}
		result = math.Sqrt(operands[0])
	case Percentage:
		result = operands[0] / 100
	}

	if math.IsNaN(result) || math.IsInf(result, 0) {
		return 0, ErrUndefinedResult
	}
	return result, nil
}

func operandCount(operation Operation) (int, bool) {
	switch operation {
	case Add, Subtract, Multiply, Divide, Power:
		return 2, true
	case SquareRoot, Percentage:
		return 1, true
	default:
		return 0, false
	}
}
