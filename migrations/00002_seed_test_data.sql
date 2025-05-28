-- +goose Up
INSERT INTO textures (id, name, price_per_dm2, image_url, in_stock) VALUES
('11111111-1111-1111-1111-111111111111', 'Натуральная кожа', 25.00, 'https://example.com/tex1.jpg', true),
('22222222-2222-2222-2222-222222222222', 'Искусственная кожа', 15.50, 'https://example.com/tex2.jpg', true),
('33333333-3333-3333-3333-333333333333', 'Замша', 30.00, 'https://example.com/tex3.jpg', false);

-- +goose Down
DELETE FROM textures WHERE id IN (
  '11111111-1111-1111-1111-111111111111',
  '22222222-2222-2222-2222-222222222222',
  '33333333-3333-3333-3333-333333333333'
);