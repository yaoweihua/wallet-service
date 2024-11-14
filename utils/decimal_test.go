// utils/validate_decimal_test.go
package utils

import (
    "testing"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
)

func TestValidateDecimal(t *testing.T) {
    tests := []struct {
        name   string
        input  decimal.Decimal
        expect bool
    }{
        {
            name:   "Valid positive decimal",
            input:  decimal.NewFromFloat(100.50),
            expect: true,
        },
        {
            name:   "Zero decimal",
            input:  decimal.NewFromFloat(0),
            expect: false,
        },
        {
            name:   "Negative decimal",
            input:  decimal.NewFromFloat(-50.75),
            expect: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidateDecimal(tt.input)
            assert.Equal(t, tt.expect, result)
        })
    }
}
