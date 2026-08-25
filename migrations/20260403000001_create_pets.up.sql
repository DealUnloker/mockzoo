CREATE TABLE IF NOT EXISTS pets (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL,
    species VARCHAR(20) NOT NULL,
    breed TEXT,
    photo_url TEXT,
    tags TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pets_status_check CHECK (status IN ('available', 'pending', 'sold')),
    CONSTRAINT pets_species_check CHECK (species IN ('dog', 'cat', 'bird', 'fish', 'other'))
);

CREATE INDEX idx_pets_status ON pets(status);
