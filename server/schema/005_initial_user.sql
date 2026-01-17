-- Create zach_monroe user (password: ZachMonroe2024!)
INSERT INTO users (username, email, password_hash, is_active)
VALUES ('zach_monroe', 'zach@example.com', '$2a$12$KEgEpwBcNCXZhLHAAmdEI.Ifo1BR88Ifj73.ENdOQgBIrPNRKyosi', true);

-- Assign all existing quotes to zach_monroe
UPDATE quotes
SET user_id = (SELECT id FROM users WHERE username = 'zach_monroe')
WHERE user_id IS NULL OR user_id NOT IN (SELECT id FROM users);
