CREATE TABLE IF NOT EXISTS orders (
    id         uuid        PRIMARY KEY,
    pastry     text        NOT NULL,
    status     text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    baked_at   timestamptz
);

CREATE INDEX IF NOT EXISTS orders_status_idx ON orders (status);
