-- +migrate Up
INSERT INTO categories (name, slug, description) VALUES
    ('Elektronik', 'elektronik', 'Produk elektronik dan gadget'),
    ('Pakaian', 'pakaian', 'Pakaian pria dan wanita'),
    ('Makanan', 'makanan', 'Makanan dan minuman ringan')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO products (category_id, name, slug, price, stock) VALUES
    (1, 'Smartphone X', 'smartphone-x', 5000000, 50),
    (1, 'Laptop Pro', 'laptop-pro', 15000000, 20),
    (2, 'Kaos Polos', 'kaos-polos', 75000, 200),
    (3, 'Kopi Arabika', 'kopi-arabika', 45000, 500)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- +migrate Down
DELETE FROM products WHERE slug IN ('smartphone-x', 'laptop-pro', 'kaos-polos', 'kopi-arabika');
DELETE FROM categories WHERE slug IN ('elektronik', 'pakaian', 'makanan');
