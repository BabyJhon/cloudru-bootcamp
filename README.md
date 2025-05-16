# Proxy server with load balancer and rate limiting
Тестовое задание для Cloud.ru Bootcamp

## Реализованный функционал
- Round-robin балансировщик
- Rate Limiter с использованием алгоритма TokenBucket
- Логирование входящих запросов, перенаправлений и ошибок
- Настройка конфигурации Rate Limiter через файл config.yml
- Rest методы для взаимодействия с клиентами
- Graceful Shutdown
- Интеграционные тесты

## Запуск проекта
В корне проекта
```bash
go mod download
go run ./cmd/proxy/main.go
```
Для запуска тестовых серверов 
```bash
go run ./cmd/test_servers/main.go
```
Для запуска тестов
```bash
go test -v ./internal/handler
```

## Api эндпоинты для работы с клиентами
- ```GET /api/ratelimit/clients```получение лимитов всех клиентов
- ```POST /api/ratelimit/clients```создание клиента
- ```GET /api/ratelimit/clients/{clientID}```получение лимитов клиента
- ```PUT /api/ratelimit/clients/{clientID}```обновление лимитов клиента
- ```DELETE /api/ratelimit/clients/{clientID}```удаление клиента
- ```GET /api/ratelimit/clients/{clientID}/tokens```получение доступных в данный момент токенов у клиента
 
