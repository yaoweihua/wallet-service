# Wallet Service

A simple wallet service API that supports user account management, balance query, deposits, withdrawals, and transfers. This project is written in Go and uses PostgreSQL and Redis for data storage and caching. It provides RESTful API interfaces along with tests by William, way1910@gmail.com.

## Table of Contents
- [Project Structure](#project-structure)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Installation & Running](#installation--running)
  - [Prerequisites](#prerequisites)
  - [Installation Steps](#installation-steps)
- [API Usage](#api-usage)
  - [Base Routes](#base-routes)
  - [Example Requests](#example-requests)
- [Testing](#testing)
  - [Run Tests](#run-tests)
  - [Mock Testing](#mock-testing)
- [TODO](#todo)
- [License](#license)

## Project Structure

```plaintext
wallet-service/
├── api/                   # API routes
│   └── routes.go          # Route initialization
├── config/                # Configuration files
│   └── config.go          # Configuration setup
├── db/                    # Database connection and initialization
│   ├── init.sql           # Database schema setup
│   ├── postgres.go       # PostgreSQL connection setup
│   └── redis.go          # Redis connection setup
├── e2e/                    # Database connection and initialization
│   ├── wallet_api_test.go  # E2E tests, testing the main scenarios and edge cases.
├── handler/               # API route handlers
│   ├── deposit.go         # Deposit request handler
│   ├── get_balance.go     # Get balance request handler
│   ├── get_transactions.go # Get transactions request handler
│   ├── withdraw.go        # Withdrawal request handler
│   └── transfer.go        # Transfer request handler
├── model/                 # Data model definitions
│   ├── transaction.go     # Transaction structure
│   └── user.go            # User structure
├── repository/            # Database operation encapsulation
│   ├── transaction_repository.go  # Transaction-related database operations
│   ├── wallet_repository.go  # Wallet-related database operations
│   ├── transaction_repository_test.go # Transaction repository tests
│   └── wallet_repository_test.go # Wallet repository tests
├── service/               # Core business logic
│   ├── deposit.go         # Deposit business logic
│   ├── get_balance.go     # Get balance business logic
│   ├── get_transactions.go # Get transactions business logic
│   ├── withdraw.go        # Withdrawal business logic
│   ├── transfer.go        # Transfer business logic
│   ├── deposit_test.go    # Deposit service tests
│   ├── get_balance_test.go # Get balance service tests
│   ├── get_transactions_test.go # Get transactions service tests
│   ├── withdraw_test.go   # Withdrawal service tests
│   └── transfer_test.go   # Transfer service tests
├── utils/                 # Utility functions
│   ├── decimal.go         # Decimal utility
│   └── logger.go          # Logger utility
├── main.go                # Main program entry point
├── docker-compose.yml     # Docker Compose configuration
├── go.mod                 # Go modules configuration
├── go.sum                 # Go modules checksum file
└── Readme.md              # Project README file
```

## Features
This wallet service implements the following basic features:
- **Deposit functionality**: Users can deposit money into their wallets.
- **Withdrawal functionality**: Users can withdraw money from their wallets.
- **Balance query**: Users can check their current wallet balance.
- **Transfer functionality**: Users can transfer money between accounts.
- **Transaction record query**: Users can view their transaction history.

## Tech Stack
- **Go**: Server-side development language.
- **PostgreSQL**: Main database for storing user accounts and transaction records.
- **Redis**: Used to cache user balances to speed up read operations.
- **Docker-Compose**: For containerizing the service.

## Installation & Running

### Prerequisites
- Go 1.17+
- PostgreSQL 13+
- Redis 6+
- Docker-Compose

### Installation Steps

1. Clone the repository:

    ```bash
    git clone https://github.com/yaoweihua/wallet-service.git
    cd wallet-service
    ```

2. Set up environment variables (optional): If the environment variables are not set, the default configurations in config/config.go will be used.
   
   Copy `.env.example` to `.env` and configure your PostgreSQL and Redis connection details.

3. Start PostgreSQL and Redis services:
   
   You can use Docker Compose or manually start PostgreSQL and Redis. The following command uses Docker Compose:

    ```bash
    docker-compose up -d
    ```

4. When the Postgres container starts, it will automatically execute db/init.sql to create the users and transactions tables, and insert some sample data. The contents are as follows:

    ```
    -- init.sql

    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        phone VARCHAR(20) UNIQUE NOT NULL,
        balance DECIMAL(10, 2) DEFAULT 0.00,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'suspended'))
    );

    INSERT INTO users (id, name, email, phone, balance, status) 
    VALUES (1, 'Alice', 'alice@example.com', '13300000001', 10.05, 'active');
    INSERT INTO users (id, name, email, phone, balance, status) 
    VALUES (2, 'Bob', 'bob@example.com', '13300000002', 50.35, 'active');
    INSERT INTO users (id, name, email, phone, balance, status) 
    VALUES (3, 'John', 'john@example.com', '13300000003', 70.55, 'active');

    CREATE TABLE IF NOT EXISTS transactions (
        id SERIAL PRIMARY KEY,
        from_user_id INT NOT NULL,  -- The user ID of the transaction initiator
        to_user_id INT,  -- The user ID of the recipient of the transaction (0 for deposits and withdrawals)
        amount DECIMAL(20, 8) NOT NULL,  -- The transaction amount, using DECIMAL type to avoid floating-point precision issues
        transaction_type VARCHAR(50) NOT NULL CHECK (transaction_type IN ('deposit', 'withdraw', 'transfer')),  -- The transaction type, restricted to deposit, withdraw, and transfer
        transaction_status VARCHAR(50) NOT NULL CHECK (transaction_status IN ('completed', 'failed')),  -- The transaction status, currently supporting completed and failed, with potential for additional intermediate statuses in the future.
        transaction_fee DECIMAL(20, 8) DEFAULT 0.00,  -- The transaction status, currently supporting completed and failed, with potential for additional intermediate statuses in the future.
        payment_method VARCHAR(50) NOT NULL,  -- The payment method, such as credit_card, bank_transfer, paypal, etc.
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, transaction_status, transaction_fee, payment_method) 
    VALUES (1, 2, 150.75, 'transfer', 'completed', 0.00, 'bank_transfer');
    INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, transaction_status, transaction_fee, payment_method) 
    VALUES (1, 0, 100.00, 'deposit', 'completed', 0.00, 'credit_card');
    ```

5. Start the service:
   
   Run the following command in the project root directory:

    ```bash
    go run main.go
    ```

   The service will be available at `http://localhost:8080`.

## API Usage

### Base Routes
The base route for the API is `/api/v1`. Here are some of the main API routes:

- `POST /v1/wallet/deposit` - Deposit
- `POST /v1/wallet/withdraw` - Withdraw
- `POST /v1/wallet/transfer` - Transfer
- `GET /v1/wallet/:user_id/balance` - Query balance
- `GET /v1/wallet/:user_id/transactions` - Get transaction records

### Example Requests

**Deposit**
- Request:  http://localhost:8080/v1/wallet/deposit
    ```json
    {
        "user_id": 1,
        "amount":  100.05
    }
    ```
- Response:
    ```json
    {
        "status": 200,
        "data": "",
        "errmsg": "Deposit successful"
    }
    ```

**Withdraw**
- Request:  http://localhost:8080/v1/wallet/withdraw
    ```json
    {
        "user_id": 1,
        "amount":  1.50
    }
    ```
- Response:
    ```json
    {
        "status": 200,
        "data": "",
        "errmsg": "Withdraw successful"
    }
    ```

**Transfer**
- Request:  http://localhost:8080/v1/wallet/transfer
    ```json
    {
        "from_user_id": 1,
        "to_user_id": 2,
        "amount":  2.05
    }
    ```
- Response:
    ```json
    {
        "status": 200,
        "data": "",
        "errmsg": "Transfer successful"
    }
    ```

**Balance Query**
- Request:  http://localhost:8080/v1/wallet/1/balance

- Response:
    ```json
    {
        "status": 200,
        "data": {
            "user_id": 1,
            "balance": "208.65"
        },
        "errmsg": ""
    }
    ```

**Get transaction records**
- Request:  http://localhost:8080/v1/wallet/1/transactions

- Response:
    ```json
    {
        "status": 200,
        "data": {
            "user_id": 1,
            "transactions": [
                {
                    "id": 5,
                    "from_user_id": 1,
                    "amount": "1.5",
                    "transaction_type": "withdraw",
                    "transaction_status": "completed",
                    "transaction_fee": "0",
                    "payment_method": "",
                    "created_at": "2024-11-12T18:22:43.954443Z",
                    "updated_at": "2024-11-12T18:22:43.954443Z"
                },
                {
                    "id": 4,
                    "from_user_id": 1,
                    "amount": "100.05",
                    "transaction_type": "deposit",
                    "transaction_status": "completed",
                    "transaction_fee": "0",
                    "payment_method": "",
                    "created_at": "2024-11-12T18:21:33.97545Z",
                    "updated_at": "2024-11-12T18:21:33.97545Z"
                }
            ]
        },
        "errmsg": ""
    }
    ```


## Testing

The project includes unit tests using the `go test` tool. The main testing files are located under the `service` and `repository` directories.

### Run Tests

To run the tests, execute the following command in the root directory:

```bash
go test ./... -race -cover
```

    ```
    Currently, unit tests are primarily focused on the service and repository layers:
        github.com/yaoweihua/wallet-service/repository  1.856s  coverage: 82.4% of statements
        github.com/yaoweihua/wallet-service/service 2.644s  coverage: 81.0% of statements
        github.com/yaoweihua/wallet-service/utils   2.212s  coverage: 92.3% of statements
    ```

## Mock Testing
The project uses mock testing to simulate database and Redis operations, ensuring that the tests do not rely on external services. This helps in creating isolated tests that focus on the business logic without the need for actual database or Redis connections.

## TODO
- This service currently does not include account management or login authentication features. By default, only 3 users are initialized. You can manually add more users by calling INSERT INTO users for now.
- The database reserves fields such as user status, transaction status, transaction fees, and payment method, which can be expanded later based on business requirements.
- For user transaction history, the search functionality may need to be enhanced in the future, depending on business needs.


## License
This project is licensed under the MIT License. See the LICENSE file for more details.
