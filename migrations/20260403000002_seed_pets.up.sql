-- Seed rows mirror internal/repo/persistent/pet/seed.go: keep both in sync.
INSERT INTO pets (id, name, status, species, breed, photo_url, tags, created_at) VALUES
    (1, 'Barsik', 'available', 'cat', 'domestic shorthair', NULL, '{fluffy,demo}', now()),
    (2, 'Rex', 'pending', 'dog', 'german shepherd', NULL, '{good-boy,demo}', now()),
    (3, 'Kesha', 'sold', 'bird', 'budgerigar', NULL, '{talks,demo}', now())
ON CONFLICT (id) DO NOTHING;

-- User-created pets start at 1001 and never collide with the stable seed ids.
SELECT setval('pets_id_seq', 1000, true);
