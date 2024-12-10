package e2e

import (
    "os"
    "testing"
    "github.com/go-resty/resty/v2"
    "github.com/stretchr/testify/assert"
    "github.com/yaoweihua/wallet-service/config"
    "github.com/yaoweihua/wallet-service/db"
    "github.com/yaoweihua/wallet-service/utils"
    "net/http"
    "github.com/jmoiron/sqlx"
    "log"
    "encoding/json"
)

// isE2ETest checks if E2E_TEST environment variable is set to "true" by William, way1910@gmail.com.
func isE2ETest() bool {
    return os.Getenv("E2E_TEST") == "true"
}

// init initializes the database connection and resets the database before tests
func init() {
    if !isE2ETest() {
        return
    }

    // Initialize the logger
    logger := utils.GetLogger()

    // Load the configuration
    cfg := config.LoadConfig()

    // Connect to the PostgreSQL database
    dbConn, err := db.ConnectPostgres(cfg)
    if err != nil {
        logger.Fatal("PostgreSQL connection failed:", err)
    }

    // Reset database state before tests
    err = resetDatabase(dbConn)
    if err != nil {
        logger.Fatal("Failed to reset database:", err)
    }
}

// resetDatabase resets the database for E2E tests (clear transactions and set balances)
func resetDatabase(dbConn *sqlx.DB) error {
    log.Println("Starting to reset default balance of users...")

    // Clear the transactions table
    _, err := dbConn.Exec("TRUNCATE TABLE transactions RESTART IDENTITY CASCADE;")
    if err != nil {
        log.Println("Error truncating transactions table:", err)
        return err
    }

    // Reset the user balances
    _, err = dbConn.Exec("UPDATE users SET balance = 10.05 WHERE id = 1;")
    if err != nil {
        log.Println("Error updating balance for user_id = 1:", err)
        return err
    }

    _, err = dbConn.Exec("UPDATE users SET balance = 50.35 WHERE id = 2;")
    if err != nil {
        log.Println("Error updating balance for user_id = 2:", err)
        return err
    }

    log.Println("Database reset complete!")
    return nil
}

// TestDepositSuccess performs an E2E test for the deposit API
func TestDepositSuccess(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 100.05
        }`).
        Post("http://localhost:8080/v1/wallet/deposit")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode())

    expected := `{
        "status": 200,
        "data": "",
        "errmsg": "Deposit successful"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestDepositZero performs an E2E test for the deposit API
func TestDepositZero(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 0
        }`).
        Post("http://localhost:8080/v1/wallet/deposit")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Deposit amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestDepositNegativeValue performs an E2E test for the deposit API
func TestDepositNegativeValue(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 0
        }`).
        Post("http://localhost:8080/v1/wallet/deposit")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Deposit amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestWithdrawSuccess performs an E2E test for the withdraw API
func TestWithdrawSuccess(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 1.50
        }`).
        Post("http://localhost:8080/v1/wallet/withdraw")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode())

    expected := `{
        "status": 200,
        "data": "",
        "errmsg": "Withdraw successful"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestWithdrawZero performs an E2E test for the withdraw API
func TestWithdrawZero(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 0
        }`).
        Post("http://localhost:8080/v1/wallet/withdraw")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Withdraw amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestWithdrawNegativeValue performs an E2E test for the withdraw API
func TestWithdrawNegativeValue(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": -1
        }`).
        Post("http://localhost:8080/v1/wallet/withdraw")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Withdraw amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestWithdrawInsufficientBalance performs an E2E test for the withdraw API
func TestWithdrawInsufficientBalance(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "user_id": 1,
            "amount": 1000
        }`).
        Post("http://localhost:8080/v1/wallet/withdraw")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Insufficient balance"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestTransferSuccess performs an E2E test for the transfer API
func TestTransferSuccess(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "from_user_id": 1,
            "to_user_id": 2,
            "amount":  2.05
        }`).
        Post("http://localhost:8080/v1/wallet/transfer")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode())

    expected := `{
        "status": 200,
        "data": "",
        "errmsg": "Transfer successful"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestTransferZero performs an E2E test for the transfer API
func TestTransferZero(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "from_user_id": 1,
            "to_user_id": 2,
            "amount":  0
        }`).
        Post("http://localhost:8080/v1/wallet/transfer")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Transfer amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestTransferNegativeValue performs an E2E test for the transfer API
func TestTransferNegativeValue(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "from_user_id": 1,
            "to_user_id": 2,
            "amount":  -1
        }`).
        Post("http://localhost:8080/v1/wallet/transfer")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Transfer amount must be greater than zero"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestTransferInsufficientBalance performs an E2E test for the transfer API
func TestTransferInsufficientBalance(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a deposit request to user_id 1 with an amount of 100.05
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        SetBody(`{
            "from_user_id": 1,
            "to_user_id": 2,
            "amount":  1000
        }`).
        Post("http://localhost:8080/v1/wallet/transfer")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

    expected := `{
        "status": 400,
        "data": "",
        "errmsg": "Insufficient balance"
    }`

    assert.JSONEq(t, expected, resp.String())
}

// TestBalanceQuery performs an E2E test for the balance query API
func TestBalanceQuery(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a balance query request to user_id 1
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        Get("http://localhost:8080/v1/wallet/1/balance")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode())

    expected := `{
        "status": 200,
        "data": {
            "user_id": 1,
            "balance": "106.55"
        },
        "errmsg": ""
    }`

    assert.JSONEq(t, expected, resp.String())
}

// Transaction represents the structure of a transaction (without `created_at` and `updated_at`)
type Transaction struct {
    ID                int     `json:"id"`
    FromUserID        int     `json:"from_user_id"`
    Amount            string  `json:"amount"`
    TransactionType   string  `json:"transaction_type"`
    TransactionStatus string  `json:"transaction_status"`
    TransactionFee    string  `json:"transaction_fee"`
    PaymentMethod     string  `json:"payment_method"`
}

// Data represents the data field in the response
type Data struct {
    UserID       int           `json:"user_id"`
    Transactions []Transaction `json:"transactions"`
}

// Response represents the entire response body
type Response struct {
    Status  int    `json:"status"`
    Data    Data   `json:"data"`
    Errmsg  string `json:"errmsg"`
}

// TestTransactionsQuery performs an E2E test for the transactions query API
func TestTransactionsQuery(t *testing.T) {
    if !isE2ETest() {
        t.Skip("Skipping E2E test, E2E_TEST is not set")
        return
    }

    client := resty.New()

    // Make a transactions query request to user_id 1
    resp, err := client.R().
        SetHeader("Content-Type", "application/json").
        Get("http://localhost:8080/v1/wallet/1/transactions")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode())

    // Define the expected response without the `created_at` and `updated_at` fields
    expected := Response{
        Status: 200,
        Data: Data{
            UserID: 1,
            Transactions: []Transaction{
                {
                    ID:                3, // Updated ID
                    FromUserID:        1,
                    Amount:            "2.05", // Transfer amount
                    TransactionType:   "transfer",
                    TransactionStatus: "completed",
                    TransactionFee:    "0",
                    PaymentMethod:     "",
                },
                {
                    ID:                2,
                    FromUserID:        1,
                    Amount:            "1.5",
                    TransactionType:   "withdraw",
                    TransactionStatus: "completed",
                    TransactionFee:    "0",
                    PaymentMethod:     "",
                },
                {
                    ID:                1,
                    FromUserID:        1,
                    Amount:            "100.05",
                    TransactionType:   "deposit",
                    TransactionStatus: "completed",
                    TransactionFee:    "0",
                    PaymentMethod:     "",
                },
            },
        },
        Errmsg: "",
    }

    // Use json.Unmarshal to decode the response body into the Response struct
    var actual Response
    err = json.Unmarshal(resp.Body(), &actual)
    assert.NoError(t, err)

    // Compare the actual response with the expected response
    assert.Equal(t, expected, actual)
}