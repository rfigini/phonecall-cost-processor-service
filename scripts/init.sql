-- Drop table if it exists (optional for dev environments)
DROP TABLE IF EXISTS public.calls;

-- Create calls table
CREATE TABLE public.calls (
    call_id UUID PRIMARY KEY,
    caller TEXT NOT NULL,
    receiver TEXT NOT NULL,
    duration_in_seconds INT NOT NULL,
    start_timestamp TIMESTAMPTZ NOT NULL,
    cost NUMERIC(10, 2),
    currency TEXT,
    cost_fetch_failed BOOLEAN DEFAULT FALSE,
    refunded BOOLEAN DEFAULT FALSE,
    refund_reason TEXT
);
