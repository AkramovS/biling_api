-- Insert test users
INSERT INTO
    users (name)
VALUES ('Иван Иванов'),
    ('Мария Петрова'),
    ('Сергей Сидоров'),
    ('Анна Смирнова'),
    ('Дмитрий Козлов');

-- Insert test accounts
INSERT INTO
    accounts (name)
VALUES ('Лицевой счет 10001'),
    ('Лицевой счет 10002'),
    ('Лицевой счет 10003'),
    ('Лицевой счет 10004'),
    ('Лицевой счет 10005'),
    ('Лицевой счет 10006'),
    ('Лицевой счет 10007'),
    ('Лицевой счет 10008'),
    ('Лицевой счет 10009'),
    ('Лицевой счет 10010');

-- Link users to accounts
INSERT INTO
    users_accounts (uid, account_id)
VALUES (1, 1),
    (1, 2),
    (1, 3), -- Иван has 3 accounts
    (2, 4),
    (2, 5), -- Мария has 2 accounts
    (3, 6), -- Сергей has 1 account
    (4, 7),
    (4, 8),
    (4, 9),
    (4, 10), -- Анна has 4 accounts
    (1, 4);
-- Ivan also has access to account 4 (shared)

-- Insert auth users (password is 'password123' hashed with bcrypt)
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("password123"), 12)
INSERT INTO
    auth_users (email, password_hash)
VALUES (
        'admin@example.com',
        '$2a$12$LHqQVV7N5JqZX.9qE5Z5Vu8FYx9X8qjKQZQX5X5X5X5X5X5X5X5X.'
    ),
    (
        'user@example.com',
        '$2a$12$LHqQVV7N5JqZX.9qE5Z5Vu8FYx9X8qjKQZQX5X5X5X5X5X5X5X5X.'
    );

-- Create API access group
INSERT INTO
    groups (name, description)
VALUES (
        'api_users',
        'Users with API access to query accounts'
    );

-- Grant permission to the group
INSERT INTO
    group_permissions (group_id, resource, action)
VALUES (1, 'accounts', 'read');

-- Add admin user to the group (user '2' is NOT in the group)
INSERT INTO group_members (group_id, auth_user_id) VALUES (1, 1);