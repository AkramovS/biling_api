-- migrations/000002_seed_data.down.sql

-- Удаляем тестовые данные в обратном порядке с учётом зависимостей
DELETE FROM account_tariff_link;
DELETE FROM tariffs;
DELETE FROM users_accounts;
DELETE FROM accounts;
DELETE FROM users;

-- Удаляем системные данные
DELETE FROM system_groups;
DELETE FROM system_rights;
DELETE FROM system_group_info;
DELETE FROM system_accounts;