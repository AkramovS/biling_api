# Как работает изменение тарифа и проверка прав доступа

## 🔄 Процесс изменения тарифа (пошагово)

### 1. Запрос от клиента

**Эндпоинт:** `PATCH /v1/account-tariffs/:id`

**Пример запроса:**
```json
{
  "tariff_id": 5,
  "version": 3
}
```

### 2. Проверка прав доступа (Middleware)

**Файл:** `cmd/api/routes.go:35-36`

```go
router.HandlerFunc(http.MethodPatch, "/v1/account-tariffs/:id",
    app.requirePermission(FIDTariffsUpdate, app.changeTariffLinkHandler))
```

**Что происходит:**
1. Middleware `requirePermission` проверяет JWT токен
2. Извлекает `user_id` из токена
3. Вызывает `HasPermission(user_id, fid=3)` 
4. Если права нет → возвращает 403 Forbidden
5. Если права есть → передает управление в `changeTariffLinkHandler`

### 3. Обработка запроса

**Файл:** `cmd/api/change_tariff.go:40-110`

**Шаги:**

1. **Парсинг параметров** (строки 42-46)
   - Извлекает `id` из URL (`/v1/account-tariffs/:id`)

2. **Парсинг тела запроса** (строки 48-58)
   - Читает JSON: `tariff_id` и `version`

3. **Валидация** (строки 60-68)
   - Проверяет, что `tariff_id > 0` и `version > 0`

4. **Получение текущего пользователя** (строка 71)
   - Извлекает пользователя из контекста (добавлен middleware)

5. **Подготовка обновления** (строки 73-79)
   - Создает объект `AccountTariffLink` с:
     - `ID` - ID записи для обновления
     - `TariffID` - новый тариф
     - `Version` - ожидаемая версия (для оптимистичной блокировки)
     - `UpdatedBy` - ID пользователя, который делает изменение

6. **Обновление в БД** (строка 82)
   - Вызывает `AccountTariffLinks.Update(link)`
   - Использует **оптимистичную блокировку** (optimistic locking)

7. **Обработка результата** (строки 83-109)
   - Если успешно → возвращает обновленную запись
   - Если конфликт версий → возвращает 409 Conflict с деталями

### 4. Оптимистичная блокировка

**Файл:** `internal/data/tariff.go:141-170`

**SQL запрос:**
```sql
UPDATE account_tariff_link
SET 
    tariff_id = $1,
    version = version + 1,        -- Увеличиваем версию
    updated_at = NOW(),
    updated_by = $2
WHERE id = $3 AND version = $4    -- Обновляем только если версия совпадает
RETURNING version, updated_at
```

**Как это работает:**
- Клиент отправляет текущую `version` (например, 3)
- Сервер обновляет только если версия в БД = 3
- Если версия изменилась (другой пользователь уже обновил) → `ErrEditConflict`
- Это предотвращает перезапись изменений другого пользователя

---

## 🔐 Как работает проверка прав доступа

### Структура таблиц

```
system_accounts (пользователи)
    ↓
system_groups (user_id → group_id)
    ↓
system_group_info (информация о группах)
    ↓
system_rights (group_id → fid)
```

### SQL запрос проверки прав

**Файл:** `internal/data/groups.go:28-47`

```sql
SELECT COUNT(*)
FROM system_rights sr
INNER JOIN system_groups sg ON sg.group_id = sr.group_id
WHERE sg.user_id = $1        -- ID пользователя
  AND sr.fid = $2            -- Feature ID (3 = tariffs:update)
  AND sg.user_id > 0
```

**Что делает:**
1. Находит все группы пользователя (`system_groups`)
2. Находит все права этих групп (`system_rights`)
3. Проверяет, есть ли среди них `fid = 3` (право на обновление тарифов)
4. Возвращает количество найденных записей (> 0 = есть право)

---

## 👥 Как определить, у кого есть доступ

### Вариант 1: SQL запрос (прямой способ)

**Найти всех пользователей с правом на обновление тарифов (fid=3):**

```sql
SELECT DISTINCT 
    sa.id,
    sa.login,
    sa.name,
    sgi.name as group_name
FROM system_accounts sa
INNER JOIN system_groups sg ON sg.user_id = sa.id
INNER JOIN system_group_info sgi ON sgi.id = sg.group_id
INNER JOIN system_rights sr ON sr.group_id = sg.group_id
WHERE sr.fid = 3              -- FIDTariffsUpdate
  AND sa.is_deleted = 0
ORDER BY sa.login;
```

### Вариант 2: Через группы

**Найти все группы с правом на обновление тарифов:**

```sql
SELECT 
    sgi.id,
    sgi.name,
    sgi.description,
    COUNT(DISTINCT sg.user_id) as users_count
FROM system_group_info sgi
INNER JOIN system_rights sr ON sr.group_id = sgi.id
LEFT JOIN system_groups sg ON sg.group_id = sgi.id
WHERE sr.fid = 3
GROUP BY sgi.id, sgi.name, sgi.description;
```

**Затем найти всех пользователей в этих группах:**

```sql
SELECT 
    sa.id,
    sa.login,
    sa.name,
    sgi.name as group_name
FROM system_accounts sa
INNER JOIN system_groups sg ON sg.user_id = sa.id
INNER JOIN system_group_info sgi ON sgi.id = sg.group_id
WHERE sgi.id IN (
    SELECT DISTINCT group_id 
    FROM system_rights 
    WHERE fid = 3
)
  AND sa.is_deleted = 0;
```

### Вариант 3: Проверка конкретного пользователя

**Проверить, есть ли у пользователя право на обновление тарифов:**

```sql
SELECT COUNT(*) > 0 as has_permission
FROM system_rights sr
INNER JOIN system_groups sg ON sg.group_id = sr.group_id
WHERE sg.user_id = 1          -- ID пользователя
  AND sr.fid = 3;             -- FIDTariffsUpdate
```

---

## 📊 Константы прав доступа

**Файл:** `cmd/api/routes.go:10-14`

```go
const (
    FIDAccountsRead  int64 = 1  // Право на чтение аккаунтов
    FIDTariffsRead   int64 = 2  // Право на чтение тарифов
    FIDTariffsUpdate int64 = 3  // Право на обновление тарифов
)
```

**Важно:** Эти значения должны совпадать с `fid` в таблице `system_rights`!

---

## 🎯 Пример настройки прав доступа

### 1. Создать пользователя

```sql
INSERT INTO system_accounts (login, password, name) 
VALUES ('admin', '$2a$12$хеш_пароля', 'Администратор');
```

### 2. Создать группу

```sql
INSERT INTO system_group_info (name, description) 
VALUES ('Администраторы', 'Полный доступ к системе');
```

### 3. Назначить пользователя группе

```sql
INSERT INTO system_groups (group_id, user_id) 
VALUES (1, 1);  -- group_id=1 (Администраторы), user_id=1 (admin)
```

### 4. Назначить права группе

```sql
INSERT INTO system_rights (group_id, fid) VALUES
    (1, 1),  -- accounts:read
    (1, 2),  -- tariffs:read
    (1, 3);  -- tariffs:update
```

**Теперь пользователь `admin` (id=1) имеет все права, включая обновление тарифов!**

---

## 🔍 Отладка проблем с доступом

### Проблема: 403 Forbidden при попытке изменить тариф

**Проверьте по порядку:**

1. **Пользователь существует?**
   ```sql
   SELECT * FROM system_accounts WHERE id = ? AND is_deleted = 0;
   ```

2. **Пользователь в группе?**
   ```sql
   SELECT * FROM system_groups WHERE user_id = ?;
   ```

3. **У группы есть право fid=3?**
   ```sql
   SELECT sr.* 
   FROM system_rights sr
   INNER JOIN system_groups sg ON sg.group_id = sr.group_id
   WHERE sg.user_id = ? AND sr.fid = 3;
   ```

4. **Проверка через API (если есть эндпоинт):**
   ```go
   permissions, err := app.models.Groups.GetUserPermissions(userID)
   // Проверить, есть ли fid=3 в списке
   ```

---

## 🔄 Визуальная схема процесса

```
┌─────────────────────────────────────────────────────────────┐
│  КЛИЕНТ                                                       │
│  PATCH /v1/account-tariffs/123                               │
│  Body: { "tariff_id": 5, "version": 3 }                      │
│  Header: Authorization: Bearer <JWT_TOKEN>                      │
└───────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  MIDDLEWARE: requirePermission(FIDTariffsUpdate=3)           │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 1. Проверка JWT токена                                │  │
│  │    → Извлечение user_id из токена                     │  │
│  │                                                        │  │
│  │ 2. Проверка прав: HasPermission(user_id, fid=3)       │  │
│  │    SQL: SELECT COUNT(*) FROM system_rights ...       │  │
│  │    WHERE user_id = ? AND fid = 3                      │  │
│  │                                                        │  │
│  │ 3. Если нет прав → 403 Forbidden                      │  │
│  │    Если есть права → Продолжить                       │  │
│  └───────────────────────────────────────────────────────┘  │
└───────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  HANDLER: changeTariffLinkHandler                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 1. Парсинг: id из URL, tariff_id и version из body    │  │
│  │                                                        │  │
│  │ 2. Валидация: tariff_id > 0, version > 0             │  │
│  │                                                        │  │
│  │ 3. Получение текущего пользователя из контекста        │  │
│  │                                                        │  │
│  │ 4. Подготовка обновления:                              │  │
│  │    AccountTariffLink {                                 │  │
│  │      ID: 123,                                          │  │
│  │      TariffID: 5,                                      │  │
│  │      Version: 3,  // Ожидаемая версия                 │  │
│  │      UpdatedBy: user_id                               │  │
│  │    }                                                   │  │
│  │                                                        │  │
│  │ 5. Обновление в БД с оптимистичной блокировкой:       │  │
│  │    UPDATE account_tariff_link                          │  │
│  │    SET tariff_id = 5, version = version + 1           │  │
│  │    WHERE id = 123 AND version = 3  ← Проверка!        │  │
│  │                                                        │  │
│  │ 6. Если версия не совпала → 409 Conflict              │  │
│  │    Если успешно → 200 OK с обновленной записью        │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🔐 Схема проверки прав доступа

```
                    ┌─────────────────┐
                    │ system_accounts │
                    │  id | login     │
                    │  1  | admin     │
                    │  2  | user1     │
                    └────────┬────────┘
                             │
                             │ user_id
                             ▼
                    ┌─────────────────┐
                    │ system_groups   │
                    │ id | group_id  │
                    │     │ user_id   │
                    │  1  |    1      │ ← admin в группе 1
                    │     |    1      │
                    │  2  |    1      │ ← user1 в группе 1
                    │     |    2      │
                    └────────┬────────┘
                             │
                             │ group_id
                             ▼
                    ┌─────────────────┐
                    │ system_rights    │
                    │ group_id | fid  │
                    │    1     |  1   │ ← accounts:read
                    │    1     |  2   │ ← tariffs:read
                    │    1     |  3   │ ← tariffs:update ✓
                    │    2     |  1   │ ← только чтение
                    └─────────────────┘

Проверка: user_id=1, fid=3
→ Находит: admin в группе 1, группа 1 имеет fid=3
→ Результат: ✅ ДОСТУП РАЗРЕШЕН
```

## 📝 Резюме

**Процесс изменения тарифа:**
1. Клиент → PATCH запрос с `tariff_id` и `version`
2. Middleware → Проверка JWT и прав (fid=3)
3. Handler → Валидация, обновление с оптимистичной блокировкой
4. Ответ → Обновленная запись или ошибка конфликта

**Проверка прав:**
- Через цепочку: `system_accounts` → `system_groups` → `system_rights`
- Проверяется наличие `fid=3` в правах групп пользователя
- Если нет → 403 Forbidden

**Кто имеет доступ:**
- Пользователи, которые:
  1. Есть в `system_accounts` (и `is_deleted=0`)
  2. Назначены в группу через `system_groups`
  3. Группа имеет право `fid=3` в `system_rights`

