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

INSERT INTO users (name, email, phone, balance, status) 
VALUES ('Alice', 'alice@example.com', '13300000001', 10.05, 'active');
INSERT INTO users (name, email, phone, balance, status) 
VALUES ('Bob', 'bob@example.com', '13300000002', 50.35, 'active');
INSERT INTO users (name, email, phone, balance, status) 
VALUES ('John', 'john@example.com', '13300000003', 70.05, 'active');

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