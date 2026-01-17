-- Drop existing foreign key constraint if it exists
ALTER TABLE quotes DROP CONSTRAINT IF EXISTS quotes_user_id_fkey;

-- Add foreign key constraint after initial user is created
ALTER TABLE quotes
ADD CONSTRAINT fk_quotes_user_id
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_quotes_user_id ON quotes(user_id);
