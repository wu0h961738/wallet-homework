
CREATE TYPE coin_type AS ENUM ('BTC', 'ETH', 'ADA');

CREATE TYPE transaction_type AS ENUM ('DEPOSIT', 'WITHDRAWAL', 'TRANSFER');

CREATE TYPE transaction_status AS ENUM ('PENDING', 'DONE', 'FAILED');

CREATE TYPE direction AS ENUM ('IN', 'OUT');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coin_type coin_type NOT NULL,
    amount NUMERIC(20, 6) DEFAULT 0,
    frozen_amount NUMERIC(20, 6) DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, coin_type)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type transaction_type NOT NULL,
    status transaction_status DEFAULT 'DONE',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transaction_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    txn_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    direction direction NOT NULL,
    amount NUMERIC(20, 6) NOT NULL,
    counterparty_wallet_id UUID REFERENCES wallets(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_transaction_entries_wallet_created ON transaction_entries(wallet_id, created_at);
CREATE INDEX idx_transaction_entries_wallet_counterparty ON transaction_entries(wallet_id, counterparty_wallet_id);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_transaction_entries_txn_id ON transaction_entries(txn_id); 