-- Fix admin and user passwords
-- Update admin password hash for 'admin123'
UPDATE users SET password = '$2a$10$kxaJjLm4y0yTmH/m3ILzqOMV/hvoZvadUBP2T/4rE4544MjxFH3PC' WHERE username = 'admin';

-- Update user password hash for 'user123'
UPDATE users SET password = '$2a$10$mPvNnaCA253Kvvq894CkAulBvuOrHDjrgUH/L/IUc1oFIfoWk0in6' WHERE username = 'user';
