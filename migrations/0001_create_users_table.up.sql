CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), 
    nickname varchar(128) unique not null,
    register_date timestamp DEFAULT CURRENT_TIMESTAMP,
    last_login_date timestamp DEFAULT CURRENT_TIMESTAMP,
    picture varchar(255) not null DEFAULT '',
    role int not null DEFAULT 0
);