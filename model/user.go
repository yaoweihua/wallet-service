// User represents an individual user in the wallet system, holding details
// such as balance, status, and contact information.
package model

import (
    "time"
    "github.com/shopspring/decimal"
)

type User struct {
    ID        int             `json:"id" db:"id"`                   // User ID
    Name      string          `json:"name" db:"name"`               // User name
    Email     string          `json:"email" db:"email"`             // User email
    Phone     string          `json:"phone" db:"phone"`             // User Phone
    Balance   decimal.Decimal `json:"balance" db:"balance"`         // User balance
    CreatedAt time.Time       `json:"created_at" db:"created_at"`   // Creation time
    UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`   // Update time
    Status    string          `json:"status" db:"status"`           // User status, such as active, inactive, suspended
}