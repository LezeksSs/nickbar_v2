CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица коктейлей
CREATE TABLE IF NOT EXISTS cocktails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    picture VARCHAR(255) not null DEFAULT '',
    rating REAL DEFAULT 0,
    description TEXT,
    recipe TEXT,
    user_id UUID NOT NULL,
    approved BOOLEAN DEFAULT FALSE,
    creation_date timestamp DEFAULT CURRENT_TIMESTAMP,
    update_date timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);