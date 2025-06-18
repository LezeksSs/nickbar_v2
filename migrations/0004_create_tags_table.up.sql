CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица тегов
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    user_id UUID,
    approved BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Таблица связей коктейлей и тегов
CREATE TABLE IF NOT EXISTS cocktail_tags (
    cocktail_id UUID NOT NULL,
    tag_id UUID NOT NULL,
    PRIMARY KEY (cocktail_id, tag_id),
    FOREIGN KEY (cocktail_id) REFERENCES cocktails(id),
    FOREIGN KEY (tag_id) REFERENCES tags(id)
);