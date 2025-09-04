-- Add response_headers column to api_check_logs table
ALTER TABLE api_check_logs 
ADD COLUMN IF NOT EXISTS response_headers TEXT DEFAULT '';

-- Create index for faster cleanup queries
CREATE INDEX IF NOT EXISTS idx_api_check_logs_checked_at 
ON api_check_logs(checked_at);

-- Update existing records to have empty response_headers
UPDATE api_check_logs 
SET response_headers = '' 
WHERE response_headers IS NULL;
