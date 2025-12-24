-- migrations/000001_create_tables.down.sql

-- Удаляем таблицы в обратном порядке с учётом зависимостей
DROP TABLE IF EXISTS account_tariff_link;
DROP TABLE IF EXISTS tariffs;
DROP TABLE IF EXISTS users_accounts;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS system_groups;
DROP TABLE IF EXISTS system_rights;
DROP TABLE IF EXISTS system_group_info;
DROP TABLE IF EXISTS system_accounts;