# Golang Project

API для управления фильмами, жанрами и отзывами с системой аутентификации и ролями.

## Требования

- Go 1.24 или выше
- PostgreSQL 16 или выше
- Docker и Docker Compose (опционально, для упрощенного запуска)

## Быстрый запуск с Docker Compose

Самый простой способ запустить проект:

1. **Запустите все сервисы (база данных + API + Admin):**
   ```bash
   docker-compose up -d
   ```

2. **Миграции применяются автоматически** при запуске API сервера.

3. **Сервисы будут доступны на:**
   - API: http://localhost:8080
   - Admin панель: http://localhost:8081

## Локальный запуск (без Docker)

### 1. Установите зависимости

```bash
go mod download
```

### 2. Запустите PostgreSQL

Используйте Docker Compose только для базы данных:
```bash
docker-compose up -d db
```

Или используйте свою локальную установку PostgreSQL.

### 3. Создайте базу данных и примените миграции

```bash
# Создайте базу данных
createdb -U app app

# Примените миграции (требуется golang-migrate)
migrate -path internal/migrations -database "postgres://app:app@localhost:5432/app?sslmode=disable" up
```

### 4. Установите переменные окружения

**Windows (PowerShell):**
```powershell
$env:DB_DSN="postgres://app:app@localhost:5432/app?sslmode=disable"
$env:JWT_SECRET="your-secret-key-here-change-in-production"
$env:PORT="8080"
$env:MIGRATIONS_PATH="internal/migrations"
```

**Windows (CMD):**
```cmd
set DB_DSN=postgres://app:app@localhost:5432/app?sslmode=disable
set JWT_SECRET=your-secret-key-here-change-in-production
set PORT=8080
set MIGRATIONS_PATH=internal/migrations
```

**Linux/Mac:**
```bash
export DB_DSN="postgres://app:app@localhost:5432/app?sslmode=disable"
export JWT_SECRET="your-secret-key-here-change-in-production"
export PORT="8080"
export MIGRATIONS_PATH="internal/migrations"
```

Или создайте файл `.env` (если используете инструмент для загрузки .env файлов).

### 5. Запустите API сервер

```bash
go run cmd/api/main.go
```

Сервер будет доступен на http://localhost:8080

## API Endpoints

### Публичные endpoints (без аутентификации)

- `GET /api/v1/health` - Проверка здоровья сервиса
- `POST /api/v1/auth/register` - Регистрация нового пользователя
- `POST /api/v1/auth/login` - Вход в систему
- `GET /api/v1/genres` - Список всех жанров
- `GET /api/v1/genres/:id` - Получить жанр по ID
- `GET /api/v1/movies` - Список всех фильмов
- `GET /api/v1/movies/:id` - Получить фильм по ID
- `GET /api/v1/movies/:id/reviews` - Список отзывов к фильму
- `GET /api/v1/users/:id/reviews` - Список отзывов пользователя

### Защищенные endpoints (требуется JWT токен)

- `GET /api/v1/me` - Информация о текущем пользователе
- `PUT /api/v1/me` - Обновление профиля текущего пользователя
- `PUT /api/v1/me/password` - Изменение пароля
- `GET /api/v1/me/reviews` - Мои отзывы
- `POST /api/v1/movies/:id/reviews` - Создать отзыв к фильму
- `PUT /api/v1/reviews/:id` - Обновить отзыв
- `DELETE /api/v1/reviews/:id` - Удалить отзыв

### Admin endpoints (требуется роль admin)

- `GET /api/v1/users` - Список всех пользователей
- `GET /api/v1/users/:id` - Получить пользователя по ID
- `PUT /api/v1/users/:id` - Обновить пользователя
- `PUT /api/v1/users/:id/role` - Изменить роль пользователя
- `DELETE /api/v1/users/:id` - Удалить пользователя
- `GET /api/v1/stats` - Статистика системы
- `GET /api/v1/audit-logs` - Логи аудита
- `POST /api/v1/genres` - Создать жанр
- `PUT /api/v1/genres/:id` - Обновить жанр
- `DELETE /api/v1/genres/:id` - Удалить жанр
- `POST /api/v1/movies` - Создать фильм
- `PUT /api/v1/movies/:id` - Обновить фильм
- `DELETE /api/v1/movies/:id` - Удалить фильм

## Аутентификация

API использует JWT токены для аутентификации. После регистрации или входа вы получите токен, который нужно передавать в заголовке:

```
Authorization: Bearer <your-token>
```

## Роли пользователей

- **user** - обычный пользователь (по умолчанию)
  - Может создавать и управлять своими отзывами
  - Может просматривать фильмы и жанры
  - Может обновлять свой профиль

- **admin** - администратор
  - Все права пользователя
  - Управление пользователями
  - Управление фильмами и жанрами
  - Просмотр статистики и логов аудита

## Особенности

### Автоматические миграции

При запуске API сервера миграции базы данных применяются автоматически. Путь к миграциям настраивается через переменную окружения `MIGRATIONS_PATH`.

### Background Workers

Система включает background worker для обработки событий отзывов:
- Автоматическое обновление среднего рейтинга фильма при создании/обновлении/удалении отзыва
- Запись событий в audit log

### Audit Logs

Все действия с отзывами (создание, обновление, удаление) логируются в таблицу audit_logs для отслеживания активности пользователей.

### Rate Limiting

В API включён простой in-memory rate limiting middleware:
- лимит: 60 запросов в минуту на один IP-адрес
- при превышении возвращается `429 Too Many Requests` с телом:
  - `{"error":"rate limit exceeded"}`


### Middleware

API использует следующие middleware:
- Request ID - уникальный ID для каждого запроса
- Logger - логирование запросов
- Rate Limit - ограничение частоты запросов
- CORS - настройка CORS заголовков
- Body Limit - ограничение размера тела запроса (1MB)
- Auth - проверка JWT токена
- Role-based access control - проверка ролей для admin endpoints

## Postman Collection

В корне репозитория лежит файл `postman_collection.json` с примерными запросами:

- Auth: регистрация и логин (`/api/v1/auth/register`, `/api/v1/auth/login`)
- Genres: список и создание жанра (для admin)
- Movies: список и создание фильма (для admin)
- Reviews: создание отзыва авторизованным пользователем

Как использовать:

1. Импортируйте `postman_collection.json` в Postman.
2. В переменной `base_url` оставьте `http://localhost:8080` или измените при необходимости.
3. Зарегистрируйте пользователя и залогиньтесь, возьмите токен и заполните:
   - `{{admin_token}}` — токен admin-пользователя
   - `{{user_token}}` — токен обычного пользователя

## Использование админских запросов

### Шаг 1: Создание админа

При регистрации все пользователи создаются с ролью `user`. Чтобы получить админа, нужно создать его напрямую в базе данных:

**Через Docker:**
```bash
docker-compose exec db psql -U app -d app
```

**Или через локальный PostgreSQL:**
```bash
psql -U app -d app
```

Затем выполните SQL:
```sql
INSERT INTO users (email, username, password_hash, role)
VALUES (
  'admin@example.com',
  'admin',
  '$2a$10$YourHashedPasswordHere',
  'admin'
);
```

**Или используйте готовый хеш для пароля `admin123`:**
```sql
INSERT INTO users (email, username, password_hash, role)
VALUES (
  'admin@example.com',
  'admin',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'admin'
);
```

### Шаг 2: Получение токена админа

1. Откройте Postman и импортируйте коллекцию `postman_collection.json`
2. Выполните запрос **Login** из папки **Public**:
   - Email: `admin@example.com`
   - Password: `admin123` (или ваш пароль)
3. Скопируйте `token` из ответа
4. В коллекции откройте вкладку **Variables** (или нажмите на название коллекции → Variables)
5. Вставьте токен в переменную `admin_token`

### Шаг 3: Использование админских запросов

Все запросы из папки **Admin** автоматически используют переменную `{{admin_token}}` в заголовке `Authorization: Bearer {{admin_token}}`.

**Доступные админские endpoints:**
- **Users**: List Users, Get User, Update User, Update User Role, Delete User
- **Genres**: Create Genre, Update Genre, Delete Genre
- **Movies**: Create Movie, Update Movie, Delete Movie
- **Stats**: Get Stats, Get Audit Logs

**Важно:** Токен действителен 24 часа. Если токен истек, выполните шаг 2 снова.

## Переменные окружения

| Переменная | Описание | Обязательная | По умолчанию |
|------------|----------|--------------|--------------|
| `PORT` | Порт для API сервера | Нет | `8080` |
| `DB_DSN` | Строка подключения к PostgreSQL | Да | - |
| `JWT_SECRET` | Секретный ключ для JWT токенов | Да | - |
| `MIGRATIONS_PATH` | Путь к файлам миграций | Нет | `internal/migrations` |

## Структура проекта

```
golang-project/
├── cmd/
│   ├── api/          # API сервер
│   └── admin/        # Admin панель
├── internal/
│   ├── database/     # Подключение к БД
│   ├── handler/      # HTTP обработчики
│   ├── middleware/   # Middleware
│   ├── migrations/   # SQL миграции
│   ├── models/       # Модели данных
│   ├── repository/   # Репозитории
│   ├── router/       # Роутинг
│   └── service/      # Бизнес-логика
└── pkg/              # Публичные пакеты
```

## Тестирование

Для запуска тестов:

```bash
go test ./...
```

Для запуска тестов с покрытием:

```bash
go test -cover ./...
```

## Остановка сервисов

```bash
docker-compose down
```

Для удаления данных базы данных:
```bash
docker-compose down -v
```
