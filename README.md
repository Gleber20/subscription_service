Subscription Service — REST API для управления пользовательскими подписками и расчёта их стоимости за заданный период.
Проект выполнен в рамках тестового задания.

Features:
- CRUD операции для подписок
- Частичное обновление (PATCH) с корректной обработкой nullable полей
- Фильтрация и пагинация списка подписок
- Расчёт общей стоимости подписок за период (помесячно)
- PostgreSQL + SQLx
- Миграции базы данных
- Swagger-документация
- Логирование HTTP-запросов

Сервис разбит по слоям, каждый отвечает за разный функционал:
- HTTP layer — обработка запросов, валидация входных данных
- Service layer — бизнес-логика и правила предметной области
- Repository layer — работа с базой данных
- Domain — сущности, DTO, фильтры и утилиты

Tech Stack
- Go
- Gin
- PostgreSQL
- sqlx
- Swagger (swaggo)
- Docker / Docker Compose
- golang-migrate
- Makefile

Все настройки в данном проекте задаются через переменные окружения (".env" файл). Образец данного файла лежит в корне проекта.

Проект обернут в контейнер и в совокупности с Makefile - это позволяет быстро его поднять и запустить.
Порядок запуска проекта следующий:

1. Запуск PostgreSQL через (make up)
2. Применение миграций через (make migrate-up)
3. Запуск приложения (make run)
4. REST API доступен по адресу: http://localhost:8080
5. Swagger UI доступен по адресу: http://localhost:8080/swagger/index.html
6. Генерация документации (make swagger)

Небольшое API Overview для наглядности:

- Create subscription (POST /api/v1/subscriptions)
- Get subscription by ID (GET /api/v1/subscriptions/{id})
- Update subscription (PATCH) (PATCH /api/v1/subscriptions/{id})
- Delete subscription (DELETE /api/v1/subscriptions/{id})
- List subscriptions (GET /api/v1/subscriptions)
- Calculate total cost (GET /api/v1/subscriptions/total?from=MM-YYYY&to=MM-YYYY)


Validation & Error Handling

- Проверка входных данных
- Корректные HTTP-статусы (400, 404, 500)
- Единый формат ошибок
- Защита от некорректных диапазонов дат

Проект не использует аунтефикацию, однако готов к расширению и её можно добавить :)