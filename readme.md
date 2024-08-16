# Сервис аутентификации пользователей

## Описание

Сервис предназначен для аутентификации пользователей. 
Пользователь вводит GUID и пароль, сервис проверяет их наличие в базе данных и возвращает ACCESS, REFRESH токены, 
которые пользователь должен использовать для доступа к другим сервисам.

## Технологии

- Go
- JWT
- PostgreSQL
- Swagger
- Docker

## Настройка

Для работы сервиса необходимо создать файл `.env` в корне проекта и заполнить его следующими переменными:

```env
POSTGRES_DB=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgrespw
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DATABASE_ADDRESS="postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable"

SERVER_PORT=8080
SERVER_ADDRESS=:$SERVER_PORT

SECRET_KEY=secret

EMAIL_HOST=smtp.mail.ru
EMAIL_PORT=587
EMAIL_USERNAME=example@mail.ru # Change this to your email
EMAIL_PASSWORD=password # Change this to your email password
```

## Запуск

Для запуска проекта необходимо выполнить следующие команды:

```bash
docker-compose up -d
```

## Swagger

После запуска проекта, документация по API ( [docs/swagger.json](docs/swagger.json) / [docs/swagger.yaml](docs/swagger.yaml) ) будет доступна по адресу:

```http
GET http://localhost:8080/api/swagger/index.html
```

## API

### Регистрация пользователя

```http
POST http://localhost:8080/api/register
```

#### Тело запроса

```json
{
  "email": "email@mail.ru",
  "password": "PASSWORD"
}
```

### Аутентификация пользователя

```http
POST http://localhost:8080/api/auth
```

#### Тело запроса

```json
{
  "guid": "GUID",
  "password": "PASSWORD"
}
```

### Обновление токенов

```http
GET http://localhost:8080/api/refresh
```

#### Заголовки

```http
Authorization-Access: ACCESS_TOKEN
Authorization-Refresh: REFRESH_TOKEN
```