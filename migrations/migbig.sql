-- Пользователи
CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- Заказы
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('created', 'pending_payment', 'paid', 'cancelled', 'shipped')),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    payment_due_at TIMESTAMP NOT NULL,         -- крайний срок оплаты
    paid_at TIMESTAMP,                         -- когда оплачен
    cancelled_at TIMESTAMP,                    -- когда отменён
    cancel_reason TEXT,                        -- причина отмены
    workflow_id TEXT,                          -- Temporal Workflow ID
    run_id TEXT                                 -- Temporal Run ID
);

-- Товары в заказе
CREATE TABLE order_items (
    id UUID PRIMARY KEY,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_name TEXT NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    price NUMERIC(10, 2) NOT NULL CHECK (price >= 0)
);

-- История смены статусов заказов
CREATE TABLE order_status_history (
    id SERIAL PRIMARY KEY,
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    reason TEXT,
    changed_at TIMESTAMP DEFAULT now()
);
