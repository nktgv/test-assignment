-- Создание таблиц для балансировщика
CREATE TABLE IF NOT EXISTS backend (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL UNIQUE,
    is_alive BOOLEAN DEFAULT TRUE,
    active_conns INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для rate-limiting
CREATE TABLE IF NOT EXISTS client (
    id SERIAL PRIMARY KEY,
    capacity INTEGER NOT NULL,
    rate_per_sec INTEGER NOT NULL,
    tokens INTEGER NOT NULL,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_backends_active ON backend(is_alive);
CREATE UNIQUE INDEX IF NOT EXISTS idx_backends_url ON backend(url);