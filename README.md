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

### P.S. Разогрев
1. Сервис аутентификации с использованием Access и Refresh токенов, обеспечивающий безопасное обновление сессий, защиту от компрометации и выдерживающий высокую нагрузку.
2. Сервис,  разработанный для учебной практики, отказал во время демонстрации преподавателям. Для решения посмотрел логи и локализовал проблему - скрипт зависал при обработке больших CSV файлов, заменил загрузку всего файла на потоковую обработку.
3. Ожидаю возможности поработать в команде опытных разработчиков над большим проектом, вырасти как специалист, получить фидбек от ментора и получить оффер в команду.
    
