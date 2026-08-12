package calculator

import (
	"errors"
	"math"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name      string
		operation Operation
		operands  []float64
		want      float64
		wantErr   error
	}{
		{name: "addition", operation: Add, operands: []float64{12.5, 7.5}, want: 20},
		{name: "subtraction", operation: Subtract, operands: []float64{5, 8}, want: -3},
		{name: "multiplication", operation: Multiply, operands: []float64{2.5, 4}, want: 10},
		{name: "division", operation: Divide, operands: []float64{9, 4}, want: 2.25},
		{name: "power", operation: Power, operands: []float64{2, 8}, want: 256},
		{name: "square root", operation: SquareRoot, operands: []float64{81}, want: 9},
		{name: "percentage", operation: Percentage, operands: []float64{17.5}, want: 0.175},
		{name: "division by zero", operation: Divide, operands: []float64{4, 0}, wantErr: ErrDivisionByZero},
		{name: "negative square root", operation: SquareRoot, operands: []float64{-1}, wantErr: ErrNegativeSquareRoot},
		{name: "invalid power domain", operation: Power, operands: []float64{-1, 0.5}, wantErr: ErrUndefinedResult},
		{name: "overflow", operation: Multiply, operands: []float64{math.MaxFloat64, 2}, wantErr: ErrUndefinedResult},
		{name: "wrong arity", operation: Add, operands: []float64{1}, wantErr: ErrInvalidOperandCount},
		{name: "unknown operation", operation: Operation("modulo"), operands: []float64{1, 2}, wantErr: ErrUnknownOperation},
		{name: "non-finite operand", operation: Add, operands: []float64{math.Inf(1), 1}, wantErr: ErrInvalidOperand},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Calculate(test.operation, test.operands)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("Calculate() error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && math.Abs(got-test.want) > 1e-12 {
				t.Errorf("Calculate() = %v, want %v", got, test.want)
			}
		})
	}
}
