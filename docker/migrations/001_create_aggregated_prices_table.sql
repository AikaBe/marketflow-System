CREATE TABLE IF NOT EXISTS aggregated_prices (
    id SERIAL PRIMARY KEY,
    pair_name VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    average_price DOUBLE PRECISION NOT NULL,
    min_price DOUBLE PRECISION NOT NULL,
    max_price DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_aggregated_prices_pair_time ON aggregated_prices (pair_name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_aggregated_prices_exchange_time ON aggregated_prices (exchange, timestamp DESC);
