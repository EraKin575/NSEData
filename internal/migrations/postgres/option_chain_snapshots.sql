CREATE TABLE IF NOT EXISTS option_chain_snapshots (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    expiry_date TIMESTAMPTZ NOT NULL,
    strike_price NUMERIC NOT NULL,
    underlying_value NUMERIC NOT NULL,
    
    ce_oi NUMERIC,
    ce_ch_oi NUMERIC,
    ce_vol NUMERIC,
    ce_iv NUMERIC,
    ce_ltp NUMERIC,
    
    pe_oi NUMERIC,
    pe_ch_oi NUMERIC,
    pe_vol NUMERIC,
    pe_iv NUMERIC,
    pe_ltp NUMERIC,
    
    intraday_pcr NUMERIC,
    pcr NUMERIC
);

CREATE INDEX IF NOT EXISTS idx_option_chain_snapshots_timestamp
    ON option_chain_snapshots(timestamp);