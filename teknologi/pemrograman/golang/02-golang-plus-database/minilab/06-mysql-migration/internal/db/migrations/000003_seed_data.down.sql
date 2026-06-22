-- +migrate Down
DELETE FROM products WHERE slug IN ('smartphone-x', 'laptop-pro', 'kaos-polos', 'kopi-arabika');
DELETE FROM categories WHERE slug IN ('elektronik', 'pakaian', 'makanan');
