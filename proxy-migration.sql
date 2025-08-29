-- Add proxy table and update endpoints table to support proxy

-- Create proxy table
CREATE TABLE IF NOT EXISTS proxies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL CHECK (port > 0 AND port <= 65535),
    username VARCHAR(255),
    password VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add proxy_id column to api_endpoints table
ALTER TABLE api_endpoints 
ADD COLUMN IF NOT EXISTS proxy_id INTEGER REFERENCES proxies(id) ON DELETE SET NULL;

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_api_endpoints_proxy_id ON api_endpoints(proxy_id);
CREATE INDEX IF NOT EXISTS idx_proxies_active ON proxies(is_active);

-- Insert some example proxies (optional)
INSERT INTO proxies (name, host, port, username, password, is_active) VALUES
('Dev Proxy', '127.0.0.1', 8888, 'admin', 'password123', true),
('Production Proxy', 'proxy.company.com', 3128, 'prod_user', 'prod_pass', true)
ON CONFLICT (name) DO NOTHING;
