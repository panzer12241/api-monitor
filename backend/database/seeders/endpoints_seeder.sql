-- API Endpoints Seeder
-- This file contains sample API endpoints for monitoring

-- Clear existing endpoints first
DELETE FROM api_endpoints;

-- Insert sample API endpoints for monitoring
INSERT INTO api_endpoints (name, url, method, headers, body, timeout_seconds, check_interval_seconds, is_active, created_at, updated_at) VALUES

-- Public APIs for testing
('JSONPlaceholder Posts', 'https://jsonplaceholder.typicode.com/posts', 'GET', '{"Content-Type": "application/json"}', '', 30, 300, true, NOW(), NOW()),
('JSONPlaceholder Users', 'https://jsonplaceholder.typicode.com/users', 'GET', '{"Content-Type": "application/json"}', '', 30, 600, true, NOW(), NOW()),
('HTTPBin GET Test', 'https://httpbin.org/get', 'GET', '{"User-Agent": "API-Monitor/1.0"}', '', 30, 300, true, NOW(), NOW()),
('HTTPBin POST Test', 'https://httpbin.org/post', 'POST', '{"Content-Type": "application/json"}', '{"test": "data", "timestamp": "2025-09-04"}', 30, 600, true, NOW(), NOW()),

-- Status Check APIs
('Google Health Check', 'https://www.google.com/', 'GET', '{}', '', 15, 300, true, NOW(), NOW()),
('GitHub API Status', 'https://api.github.com/status', 'GET', '{"User-Agent": "API-Monitor/1.0"}', '', 30, 300, true, NOW(), NOW()),

-- Sample REST APIs
('Cat Facts API', 'https://catfact.ninja/fact', 'GET', '{}', '', 30, 900, true, NOW(), NOW()),
('Dog API Random', 'https://dog.ceo/api/breeds/image/random', 'GET', '{}', '', 30, 600, true, NOW(), NOW()),

-- Test APIs (for demonstration)
('Local Test API (will fail)', 'http://localhost:3000/api/test', 'GET', '{"Content-Type": "application/json"}', '', 10, 300, false, NOW(), NOW()),
('Slow Response Test', 'https://httpbin.org/delay/2', 'GET', '{}', '', 5, 1800, false, NOW(), NOW())

ON CONFLICT DO NOTHING;
