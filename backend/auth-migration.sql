-- Create users table for JWT authentication
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Insert default admin user (password: admin123)
-- Password hash for 'admin123' using bcrypt cost 10
INSERT INTO users (username, email, password, role, is_active, created_at, updated_at) 
VALUES (
    'admin', 
    'admin@example.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    'admin', 
    true, 
    NOW(), 
    NOW()
) ON CONFLICT (username) DO NOTHING;

-- Insert default user (password: user123)
-- Password hash for 'user123' using bcrypt cost 10
INSERT INTO users (username, email, password, role, is_active, created_at, updated_at) 
VALUES (
    'user', 
    'user@example.com', 
    '$2a$10$M1JTa/0v2Qs9W8aO0OKOK.VdQy6FJEd3j0RGj8LDcLlJX8Z5dYX2i', -- user123
    'user', 
    true, 
    NOW(), 
    NOW()
) ON CONFLICT (username) DO NOTHING;
