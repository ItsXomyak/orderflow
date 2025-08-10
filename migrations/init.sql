-- Таблица заказов
CREATE TABLE IF NOT EXISTS orders (
    id            TEXT PRIMARY KEY,
    customer_id   TEXT        NOT NULL,
    status        TEXT        NOT NULL CHECK (status IN (
                     'pending','validating','payment','completed','failed','cancelled'
                   )),
    total_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_id    TEXT,
    failure_reason TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ
);

-- Таблица товаров заказа
CREATE TABLE IF NOT EXISTS order_items (
    id         BIGSERIAL PRIMARY KEY,
    order_id   TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    name       TEXT NOT NULL,
    quantity   INT  NOT NULL CHECK (quantity > 0),
    price      NUMERIC(12,2) NOT NULL CHECK (price >= 0)
);

-- Таблица товаров (склад)
CREATE TABLE IF NOT EXISTS products (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    sku        TEXT UNIQUE NOT NULL,
    price      NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    available  INT NOT NULL DEFAULT 0 CHECK (available >= 0),
    reserved   INT NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица резервирований
CREATE TABLE IF NOT EXISTS reservations (
    id         TEXT PRIMARY KEY,
    order_id   TEXT NOT NULL,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INT NOT NULL CHECK (quantity > 0),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица платежей
CREATE TABLE IF NOT EXISTS payments (
    id             TEXT PRIMARY KEY,
    order_id       TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    customer_id    TEXT NOT NULL,
    amount         NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency       TEXT NOT NULL,
    status         TEXT NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    payment_method TEXT NOT NULL,
    transaction_id TEXT,
    failure_reason TEXT,
    processed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Таблица уведомлений
CREATE TABLE IF NOT EXISTS notifications (
    id         TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    order_id   TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    type       TEXT NOT NULL CHECK (type IN ('order_confirmed', 'order_failed', 'order_cancelled', 'payment_failed')),
    channel    TEXT NOT NULL CHECK (channel IN ('email', 'sms', 'push')),
    status     TEXT NOT NULL CHECK (status IN ('pending', 'sent', 'failed')),
    subject    TEXT,
    message    TEXT NOT NULL,
    metadata   JSONB,
    sent_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для orders
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at  ON orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Индексы для products
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_available ON products(available);

-- Индексы для reservations
CREATE INDEX IF NOT EXISTS idx_reservations_order_id ON reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_product_id ON reservations(product_id);
CREATE INDEX IF NOT EXISTS idx_reservations_expires_at ON reservations(expires_at);

-- Индексы для payments
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);

-- Индексы для notifications
CREATE INDEX IF NOT EXISTS idx_notifications_order_id ON notifications(order_id);
CREATE INDEX IF NOT EXISTS idx_notifications_customer_id ON notifications(customer_id);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Применяем триггеры ко всем таблицам с updated_at
DO $$
BEGIN
  -- orders
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'orders_set_updated_at') THEN
    CREATE TRIGGER orders_set_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
  
  -- products
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'products_set_updated_at') THEN
    CREATE TRIGGER products_set_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
  
  -- payments
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'payments_set_updated_at') THEN
    CREATE TRIGGER payments_set_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
  
  -- notifications
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'notifications_set_updated_at') THEN
    CREATE TRIGGER notifications_set_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
  END IF;
END $$;
