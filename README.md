# **Тестовое задание для стажёра Backend-направления (зимняя волна 2025)**

## Магазин мерча

В Авито существует внутренний магазин мерча, где сотрудники могут приобретать товары за монеты (coin). Каждому новому сотруднику выделяется 1000 монет, которые можно использовать для покупки товаров. Кроме того, монеты можно передавать другим сотрудникам в знак благодарности или как подарок.

## Запуск проекта

1. Клонировать репозиторий
```bash
git clone https://github.com/gratefultolord/merch-store.git
```
2. Установить зависимости
```bash
go mod download
```
3. Заполнить файл .env
```.dotenv
DB_HOST=db
DB_PORT=5432
DB_USER=avito
DB_PASSWORD=avito
DB_NAME=avito_shop

JWT_SECRET=supersecretkey

SERVER_PORT=8080
```
4. Собрать образ
```bash
docker build -t merch-store .
```
5. Запустить docker compose
```bash
docker compose up
```
6. Сделать запрос
```bash
curl -X POST http://localhost:8080/api/auth \
     -H "Content-Type: application/json" \
     -d '{"username": "user1", "password": "password1"}'
```
**Примеры остальных запросов можно найти в schema.json и schema.yaml**
## **Общие вводные**

**Мерч** — это продукт, который можно купить за монетки. Всего в магазине доступно 10 видов мерча. Каждый товар имеет уникальное название и цену. Ниже приведён список наименований и их цены.

| Название     | Цена |
|--------------|------|
| t-shirt      | 80   |
| cup          | 20   |
| book         | 50   |
| pen          | 10   |
| powerbank    | 200  |
| hoody        | 300  |
| umbrella     | 200  |
| socks        | 10   |
| wallet       | 50   |
| pink-hoody   | 500  |

## **Тестирование проекта**
```bash
go test -cover ./...
```
## **Вопросы по заданию**
1) Статус-код 401 (Unauthorized) в /api/auth. Думаю, что для этого эндпоинта 401 не совсем актуален, поскольку пользователь авторизуется по этому запросу.
2) Надо ли было включать в coinHistory транзакции, связанные с покупкой предметов? В моем решении они не включены, так как я подумал, что там должны быть именно транзакции между пользователями.