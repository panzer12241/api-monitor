-- Create main tables for API monitoring system

-- Create api_endpoints table
CREATE TABLE IF NOT EXISTS api_endpoints (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) NOT NULL DEFAULT 'GET',
    headers JSONB DEFAULT '{}',
    body TEXT DEFAULT '',
    timeout_seconds INTEGER NOT NULL DEFAULT 30,
    check_interval_seconds INTEGER NOT NULL DEFAULT 300,
    is_active BOOLEAN NOT NULL DEFAULT true,
    proxy_id INTEGER NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create api_check_logs table
CREATE TABLE IF NOT EXISTS api_check_logs (
    id SERIAL PRIMARY KEY,
    endpoint_id INTEGER NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    response_body TEXT,
    response_headers TEXT DEFAULT '',
    error_message TEXT,
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (endpoint_id) REFERENCES api_endpoints(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_api_endpoints_active ON api_endpoints(is_active);
CREATE INDEX IF NOT EXISTS idx_api_endpoints_check_interval ON api_endpoints(check_interval_seconds);
CREATE INDEX IF NOT EXISTS idx_api_check_logs_endpoint_id ON api_check_logs(endpoint_id);
CREATE INDEX IF NOT EXISTS idx_api_check_logs_checked_at ON api_check_logs(checked_at);
CREATE INDEX IF NOT EXISTS idx_api_check_logs_status_code ON api_check_logs(status_code);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for api_endpoints
CREATE TRIGGER update_api_endpoints_updated_at BEFORE UPDATE
ON api_endpoints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
