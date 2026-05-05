ALTER TABLE products
ADD COLUMN IF NOT EXISTS popularity_score INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS avg_rating NUMERIC(3,2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS rating_count INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_products_popularity ON products (popularity_score DESC);
CREATE INDEX IF NOT EXISTS idx_products_rating ON products (avg_rating DESC);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);
