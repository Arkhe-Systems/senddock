-- +goose Up
  CREATE TABLE users (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      email VARCHAR(255) NOT NULL UNIQUE,
      password_hash VARCHAR(255),
      name VARCHAR(255) NOT NULL,

      -- plan & features
      plan VARCHAR(50) NOT NULL DEFAULT 'free',

      -- limits (only works on cloud)
      monthly_email_limit INT NOT NULL DEFAULT 2000,
      monthly_emails_sent INT NOT NULL DEFAULT 0,
      monthly_reset_at TIMESTAMPTZ NOT NULL DEFAULT date_trunc('month', NOW()) + INTERVAL '1
  month',

      -- OAuth
      provider VARCHAR(50),
      provider_id VARCHAR(255),

      -- Status
      is_verified BOOLEAN NOT NULL DEFAULT false,
      is_banned BOOLEAN NOT NULL DEFAULT false,
      ban_reason TEXT,

      -- Payments (solo cloud)
      customer_id VARCHAR(255),
      subscription_id VARCHAR(255),
      subscription_status VARCHAR(50),
      plan_changed_at TIMESTAMPTZ,

      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE INDEX idx_users_provider ON users(provider, provider_id)
      WHERE provider IS NOT NULL;

  CREATE INDEX idx_users_plan ON users(plan);

-- +goose Down
  DROP TABLE users;