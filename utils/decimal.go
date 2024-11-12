// Package utils provides utility functions and helpers, including
// decimal handling and operations, to support various components
// of the wallet service.
package utils

import "github.com/shopspring/decimal"

// ValidateDecimal ensures the decimal value is positive.
func ValidateDecimal(amount decimal.Decimal) bool {
    return amount.GreaterThan(decimal.Zero)
}