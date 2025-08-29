-- Create API endpoints table
CREATE TABLE IF NOT EXISTS api_endpoints (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) DEFAULT 'GET',
    headers JSONB DEFAULT '{}',
    body TEXT,
    timeout_seconds INTEGER DEFAULT 30,
    check_interval_seconds INTEGER DEFAULT 60,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create API check logs table (for detailed logging, supplement to Prometheus)
CREATE TABLE IF NOT EXISTS api_check_logs (
    id SERIAL PRIMARY KEY,
    endpoint_id INTEGER REFERENCES api_endpoints(id) ON DELETE CASCADE,
    status_code INTEGER,
    response_time_ms INTEGER,
    response_body TEXT,
    error_message TEXT,
    checked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_api_endpoints_active ON api_endpoints(is_active);
CREATE INDEX IF NOT EXISTS idx_api_check_logs_endpoint_id ON api_check_logs(endpoint_id);
CREATE INDEX IF NOT EXISTS idx_api_check_logs_checked_at ON api_check_logs(checked_at);

-- Insert sample data
INSERT INTO api_endpoints (name, url, method, check_interval_seconds) VALUES
('Google Homepage', 'https://www.google.com', 'GET', 60),
('JSONPlaceholder Posts', 'https://jsonplaceholder.typicode.com/posts/1', 'GET', 30);
