CREATE TABLE accounts (
    id bigint PRIMARY KEY,
    balance_cents bigint NOT NULL
);

CREATE TABLE settlement_runs (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    batch_key text NOT NULL UNIQUE
);

CREATE TABLE pending_settlements (
    batch_key text NOT NULL,
    ordinal integer NOT NULL,
    account_id bigint NOT NULL REFERENCES accounts(id),
    amount_cents bigint NOT NULL,
    run_id bigint REFERENCES settlement_runs(id),
    PRIMARY KEY (batch_key, ordinal)
);
