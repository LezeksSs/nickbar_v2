CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица номенклатуры ингредиентов
CREATE TABLE IF NOT EXISTS ingredients_nomenclature (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    picture VARCHAR(255) not null DEFAULT '',
    user_id UUID NOT NULL,
    approved BOOLEAN DEFAULT FALSE
);

-- Таблица ингредиентов
CREATE TABLE IF NOT EXISTS ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cocktail_id UUID NOT NULL,
    nomenclature_id UUID NOT NULL,
    amount REAL,
    measure VARCHAR(50),
    optional BOOLEAN DEFAULT FALSE,
    decorative BOOLEAN DEFAULT FALSE,
    position INT,
    FOREIGN KEY (cocktail_id) REFERENCES cocktails(id),
    FOREIGN KEY (nomenclature_id) REFERENCES ingredients_nomenclature(id)
);

-- Таблица замен ингредиентов
CREATE TABLE IF NOT EXISTS ingredient_replacements (
    ingredient_id UUID NOT NULL,
    replacement_id UUID NOT NULL,
    PRIMARY KEY (ingredient_id, replacement_id),
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id),
    FOREIGN KEY (replacement_id) REFERENCES ingredients(id)
);