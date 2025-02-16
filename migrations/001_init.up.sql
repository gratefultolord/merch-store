-- Создание таблицы users --
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    balance INT DEFAULT 1000 NOT NULL
);

-- Создание таблицы items --
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    price INT NOT NULL
);

-- Создание таблицы transactions --
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES users(id) ON DELETE CASCADE,
    receiver_id INT REFERENCES users(id) ON DELETE CASCADE,
    amount INT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Создание таблицы user_inventory --
CREATE TABLE IF NOT EXISTS user_inventory (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    item_id INT REFERENCES items(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    UNIQUE(user_id, item_id)
);

-- Решение вопроса с покупкой --
ALTER TABLE transactions DROP CONSTRAINT transactions_receiver_id_fkey;
ALTER TABLE transactions ALTER COLUMN receiver_id DROP NOT NULL;

INSERT INTO users (id, username, password_hash, balance)
VALUES (-1, 'system', '', 0);