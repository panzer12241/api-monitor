-- User Table Seeder
-- This file contains sample users for development and testing

-- Clear existing users first (except admin and user)
DELETE FROM users WHERE username NOT IN ('admin', 'user');

-- Insert sample users
INSERT INTO users (username, password, role, is_active, created_at, updated_at) VALUES
-- Managers
('manager1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, NOW(), NOW()), -- password: admin123

ON CONFLICT (username) DO NOTHING;
