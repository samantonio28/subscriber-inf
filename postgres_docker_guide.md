# Работа с PostgreSQL в Docker контейнере

## Подключение к контейнеру

### 1. Вход в контейнер через bash
```bash
docker-compose exec postgres bash
```

### 2. Подключение к PostgreSQL через psql
После входа в контейнер выполните:
```bash
psql -U postgres -d dev
```

Или одной командой:
```bash
docker-compose exec postgres psql -U postgres -d dev
```

## Основные команды PostgreSQL

### Просмотр списка баз данных
```sql
\l
```

### Подключение к базе данных
```sql
\c dev
```

### Просмотр списка таблиц
```sql
\dt
```

### Просмотр структуры таблицы
```sql
\d table_name
```

Например:
```sql
\d subscriptions
```

### Выход из psql
```sql
\q
```

## Примеры SQL запросов

### Просмотр всех таблиц
```sql
SELECT tablename FROM pg_tables WHERE schemaname = 'public';
```

### Просмотр данных из таблицы subscriptions
```sql
SELECT * FROM subscriptions LIMIT 10;
```

### Просмотр структуры таблицы subscriptions
```sql
\d subscriptions
```

### Примеры запросов к таблицам

1. Просмотр всех сервисов:
```sql
SELECT * FROM services;
```

2. Просмотр всех подписок:
```sql
SELECT * FROM subscriptions;
```

3. Просмотр всех пользователей:
```sql
SELECT * FROM users;
```

4. Подсчет количества подписок:
```sql
SELECT COUNT(*) FROM subscriptions;
```

5. Поиск подписок по типу:
```sql
SELECT * FROM subscriptions WHERE sub_type = 'usual';
```

## Параметры подключения

Из конфигурационного файла `configs/postgres.yaml`:

- Хост: localhost
- Порт: 8000
- Пользователь: postgres
- Пароль: secret
- База данных: dev

## Прямое подключение к PostgreSQL без входа в контейнер

Если у вас установлен клиент PostgreSQL локально:

```bash
psql -h localhost -p 8000 -U postgres -d dev
```

Пароль: secret