
-- Тестовые данные

-- Системные пользователи
INSERT INTO system_accounts (login, password, name) VALUES
    -- Пароль: admin123
    ('admin', '$2a$12$ZMoG.NBGhq6YZwWgVAzn3.lamIZbVWVed0U5NX6QdpE3u0ljCGar.', 'Администратор')     
    -- Пароль: password123
    , ('username', '$2a$12$u9/OeAiKyR74JFKJ78tFkO0lzhz76vlmlpd1c0Fzs29DcT1yhcvvm', 'НЕКИЙ ТЕСТОВЫЙ АДМИН');

-- Группы
INSERT INTO system_group_info (name, description) VALUES
    ('Администраторы', 'Полный доступ ко всем функциям системы');

-- Права для группы Администраторы (group_id=1): просмотр счетов, просмотр тарифов, изменение тарифов
INSERT INTO system_rights (group_id, fid) VALUES
    (1, 1),  -- FID 1: просмотр счетов
    (1, 2),  -- FID 2: просмотр тарифов
    (1, 3);  -- FID 3: изменение тарифов

-- Привязка пользователя admin (id=1) к группе Администраторы (group_id=1)
INSERT INTO system_groups (group_id, user_id) VALUES
    (1, 1);

-- Обычные пользователи
INSERT INTO users (name) VALUES
    ('Иван Петров'),
    ('Мария Сидорова'),
    ('Алексей Козлов'),
    ('Елена Новикова'),
    ('Дмитрий Волков');

INSERT INTO accounts (id) VALUES
    (1),
    (2),
    (3);

INSERT INTO users_accounts (uid, account_id) VALUES
    (1, 1),
    (1, 2),
    (2, 1),
    (3, 2),
    (3, 3),
    (4, 1),
    (5, 3);



insert into tariffs (name, description, price) values
    ('Тариф 1', 'Описание тарифа 1', 100.00),
    ('Тариф 2', 'Описание тарифа 2', 200.00),
    ('Тариф 3', 'Описание тарифа 3', 300.00);

INSERT INTO account_tariff_link (account_id, tariff_id) VALUES
    (1, 1),
    (2, 2),
    (3, 3);
