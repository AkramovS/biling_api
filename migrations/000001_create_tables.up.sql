-- migrations/000001_create_system_tables.up.sql

-- 1. Системные пользователи
CREATE TABLE system_accounts (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL DEFAULT '',
    password VARCHAR(255) NOT NULL DEFAULT '',
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    is_deleted INT NOT NULL DEFAULT 0
);

-- 2. Системные группы
CREATE TABLE system_group_info (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT ''
);

-- 3. Права для групп
CREATE TABLE system_rights (
    group_id INT NOT NULL DEFAULT 0,
    fid INT NOT NULL DEFAULT 0,
    UNIQUE (group_id, fid)
);

-- 4. Принадлежность пользователей к группам
CREATE TABLE system_groups (
    id SERIAL PRIMARY KEY,
    group_id INT NOT NULL DEFAULT 0,
    user_id INT NOT NULL DEFAULT 0
);


-- 5 Таблица пользователей
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- 6. Таблица аккаунтов
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY
);

-- 7. Связь пользователей с аккаунтами
CREATE TABLE users_accounts (
    id SERIAL PRIMARY KEY,
    uid INT NOT NULL,
    account_id INT NOT NULL
);


-- 8. Таблица тарифов
CREATE TABLE tariffs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 9. Связь аккаунтов с тарифами
CREATE TABLE account_tariff_link (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id),
    tariff_id INT NOT NULL,
    version  BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by INT REFERENCES system_accounts(id),
    UNIQUE(account_id)
);