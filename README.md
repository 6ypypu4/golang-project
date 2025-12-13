# Golang Project

## Требования

- Go 1.22 или выше
- PostgreSQL 16 или выше
- Docker и Docker Compose (опционально, для упрощенного запуска)

## Быстрый запуск с Docker Compose

Самый простой способ запустить проект:

1. **Запустите все сервисы (база данных + API):**
   ```bash
   docker-compose up -d
   ```

2. **Примените миграции базы данных:**
   ```bash
   # Если у вас установлен golang-migrate
   migrate -path internal/migrations -database "postgres://app:app@localhost:5432/app?sslmode=disable" up
   ```

3. **API будет доступен на:** http://localhost:8080

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
```

**Windows (CMD):**
```cmd
set DB_DSN=postgres://app:app@localhost:5432/app?sslmode=disable
set JWT_SECRET=your-secret-key-here-change-in-production
set PORT=8080
```

**Linux/Mac:**
```bash
export DB_DSN="postgres://app:app@localhost:5432/app?sslmode=disable"
export JWT_SECRET="your-secret-key-here-change-in-production"
export PORT="8080"
```

Или создайте файл `.env` (если используете инструмент для загрузки .env файлов).

### 5. Запустите API сервер

```bash
go run cmd/api/main.go
```

Сервер будет доступен на http://localhost:8080

## Переменные окружения

| Переменная | Описание | Обязательная | По умолчанию |
|------------|----------|--------------|--------------|
| `PORT` | Порт для API сервера | Нет | `8080` |
| `DB_DSN` | Строка подключения к PostgreSQL | Да | - |
| `JWT_SECRET` | Секретный ключ для JWT токенов | Да | - |

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

## Остановка сервисов

```bash
docker-compose down
```

Для удаления данных базы данных:
```bash
docker-compose down -v
```
