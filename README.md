# URL Shortener

URL Shortener — это сервис для сокращения ссылок на Go с использованием SQLite и HTTP API.

## Возможности
- Сокращение длинных URL с возможностью указать свой alias
- Редирект по короткой ссылке
- Удаление короткой ссылки
- Аутентификация через Basic Auth

## Структура проекта
- `cmd/url_shortener/main.go` — точка входа, запуск HTTP-сервера
- `internal/storage/sqlite/` — работа с SQLite
- `internal/http_server/handlers/` — обработчики HTTP-запросов (save, delete, redirect)
- `config/` — конфигурационные файлы (пример: `local.yaml`)
- `tests/` — интеграционные тесты

## Быстрый старт

1. **Клонируйте репозиторий и установите зависимости:**

```sh
git clone <repo_url>
cd url_shortener
go mod download
```

2. **Настройте конфиг:**

Пример `config/local.yaml`:
```yaml
env: "local" # local, dev, prod
storage_path: "./storage/storage.db"
http_server:
  address: "localhost:8082"
  timeout: 4s
  idle_timeout: 60s
  user: "us"
  password: "pass"
```

3. **Запустите сервер:**

```sh
go run ./cmd/url_shortener/main.go
```

## API

### Сохранить ссылку
- **POST** `/save`
- Basic Auth: `user` и `password` из конфига
- Тело запроса:
```json
{
  "url": "https://example.com",
  "alias": "myalias" // не обязательно
}
```
- Ответ:
```json
{
  "status": "OK",
  "alias": "myalias"
}
```

### Редирект по короткой ссылке
- **GET** `/{alias}`
- Basic Auth: `user` и `password`
- Ответ: 302 Redirect на оригинальный URL

### Удалить ссылку
- **DELETE** `/delete/{alias}`
- Basic Auth: `user` и `password`
- Ответ:
```json
{
  "status": "OK"
}
```

## Тесты

Для запуска интеграционных тестов:
```sh
go test ./tests
```

## Зависимости
- Go 1.24+
- SQLite (через go-sqlite3)
- [go-chi/chi](https://github.com/go-chi/chi) — роутер
- [go-playground/validator](https://github.com/go-playground/validator) — валидация

## Лицензия
MIT 