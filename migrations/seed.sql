-- Демо-товары для тестирования
INSERT INTO products (id, name, sku, price, available, reserved, created_at, updated_at) VALUES
('prod-001', 'iPhone 15 Pro', 'IPHONE-15-PRO-128', 999.99, 50, 0, NOW(), NOW()),
('prod-002', 'MacBook Air M2', 'MACBOOK-AIR-M2-256', 1199.99, 25, 0, NOW(), NOW()),
('prod-003', 'AirPods Pro', 'AIRPODS-PRO-2', 249.99, 100, 0, NOW(), NOW()),
('prod-004', 'iPad Air', 'IPAD-AIR-64', 599.99, 30, 0, NOW(), NOW()),
('prod-005', 'Apple Watch Series 9', 'APPLE-WATCH-S9-45', 399.99, 75, 0, NOW(), NOW()),
('prod-006', 'MacBook Pro 14"', 'MACBOOK-PRO-14-M3', 1999.99, 15, 0, NOW(), NOW()),
('prod-007', 'iPhone 15', 'IPHONE-15-128', 799.99, 60, 0, NOW(), NOW()),
('prod-008', 'iPad Pro 12.9"', 'IPAD-PRO-12-9-256', 1099.99, 20, 0, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
