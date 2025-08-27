# Задание

Необходимо разработать демонстрационный сервис с простейшим интерфейсом, отображающий данные о заказе.

Данное задание предполагает создание небольшого микросервиса на Go с использованием базы данных и очереди сообщений. Сервис будет получать данные заказов из очереди (Kafka), сохранять их в базу данных (PostgreSQL) и кэшировать в памяти для быстрого доступа.

Что нужно сделать:

- **Развернуть локально базу данных:**

1. Cоздать новую базу данных для сервиса

2. Настроить пользователя: заведите пользователя и выдайте права на созданную БД

3. Создать таблицы: спроектируйте структуру для хранения полученных данных о заказах, ориентируясь на прилагаемую модель данных.

- **Разработать сервис:**

1. Написать приложение на Go, реализующее описанные ниже функции.

2. Разработать простейший интерфейс для отображения полученных данных по ID заказа.

3. Подключиться и подписаться на канал сообщений: настроить получение данных из брокера сообщений (Kafka).

4. Сохранять полученные данные в БД: при приходе нового сообщения о заказе, парсить его и вставлять соответствующую запись(и) в базу данных (PostgreSQL).

5. Реализовать кэширование данных в сервисе: хранить последние полученные данные заказов в памяти (например, в map), чтобы быстро выдавать их по запросу.

6. При перезапуске восстанавливать кеш из БД: при старте сервиса заполнять кеш актуальными данными из базы, чтобы продолжить обслуживание запросов без задержек.

7. Запустить HTTP-сервер для выдачи данных по ID: реализовать HTTP-эндпоинт, который по order_id будет возвращать данные заказа из кеша (JSON API). Если в кеше данных нет, можно подтягивать из БД.

- **Разработать простой веб-интерфейс** — страницу (HTML/JS), где можно ввести ID заказа и получить информацию о нём, обращаясь к вышеописанному HTTP API.

> Примечание: модель данных заказа (поля заказа, доставки, оплаты и товаров) прилагается в JSON-файле.

Данные, приходящие из очереди, могут быть невалидными — необходимо предусмотреть обработку ошибок (например, игнорируйте или логируйте некорректные сообщения). В ходе реализации убедитесь, что при сбоях (ошибка базы, падение сервиса) данные не теряются — используйте транзакции, механизм подтверждения сообщений от брокера и т.д.

После реализации убедитесь, что:

1. Сервис подключается к брокеру сообщений (Kafka) и обрабатывает сообщения онлайн (можно написать скрипт-эмулятор отправки сообщений).

2. Кеш действительно ускоряет получение данных (например, при повторных запросах по одному и тому же ID).

3. HTTP-сервер возвращает корректные данные в формате JSON.

Пример запроса:
GET http://localhost:8081/order/<order_uid> должен вернуть JSON с информацией о заказе.

4. Интерфейс отображает данные понятным образом после ввода ID и нажатия кнопки.

**Модель данных**

```json
{
   "order_uid": "b563feb7b2b84b6test",
   "track_number": "WBILMTESTTRACK",
   "entry": "WBIL",
   "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
   },
   "payment": {
      "transaction": "b563feb7b2b84b6test",
      "request_id": "",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
   },
   "items": [
      {
         "chrt_id": 9934930,
         "track_number": "WBILMTESTTRACK",
         "price": 453,
         "rid": "ab4219087a764ae0btest",
         "name": "Mascaras",
         "sale": 30,
         "size": "0",
         "total_price": 317,
         "nm_id": 2389212,
         "brand": "Vivienne Sabo",
         "status": 202
      }
   ],
   "locale": "en",
   "internal_signature": "",
   "customer_id": "test",
   "delivery_service": "meest",
   "shardkey": "9",
   "sm_id": 99,
   "date_created": "2021-11-26T06:22:19Z",
   "oof_shard": "1"
}
```

# Решение

### Архитектура проекта на бэкенде

- Чистая архитектура с разделением на слои:

```
   cmd/            
      main/        → Точка входа для основного приложения (main)
      producer/    → Точка входа для фейкового продюсера
   internal/
      app/            → Ядро приложения
         delivery/    → GraphQL-роуты
         usecase/     → Бизнес-логика
         repository/  → Хранилища (PostgreSQL + in-memory)
         mocks/       → Моки для тестов
         models/      → Модели
      config/         → Конфигурация
      middleware/     → Мидлвари
      utils/          → Вспомогательные утилиты
      kafka/          → Логика по работе с Кафкой
   storage/        → Миграции
```

### Кэш

В качестве инструмента для кэширования и последующего быстрого доступа используется Redis.

В Redis сохраняю при получении пользователем заказа по UID. До этого момента в Кэш ничего не сохраняется. Данные в кэше хранятся 7 дней.

### Kafka

Три брокера Кафки: 1 лидер и два ведомых. Это необходимо для создания нескольких партиций и для репликации. 

Для визуализации сообщений, топиков и других полезных данных подключил контейнер kafka-ui.

### API

Получение заказа по UID

- `GET /api/v1/orders/{orderUID}`

- Успешный ответ:

```json
{
	"customer_id": "customer222",
	"date_created": "2025-08-25T10:20:35Z",
	"delivery": {
		"address": "Central st., 77",
		"city": "Saint Petersburg",
		"email": "test311@example.com",
		"name": "Sarah Johnson",
		"phone": "+79012302330",
		"region": "Moscow Oblast",
		"zip": "702701"
	},
	"delivery_service": "russian-post",
	"entry": "WBIL",
	"internal_signature": "",
	"items": [
		{
			"brand": "Huawei",
			"chrt_id": 1127547,
			"name": "Tablet",
			"nm_id": 958951,
			"price": 20000,
			"rid": "ab4219087a764ae0btest0",
			"sale": 2,
			"size": "3",
			"status": 202,
			"total_price": 15600,
			"track_number": "WBILMTESTTRACK15"
		},
		{
			"brand": "JBL",
			"chrt_id": 6264093,
			"name": "Speaker",
			"nm_id": 2134,
			"price": 7000,
			"rid": "ab4219087a764ae0btest1",
			"sale": 16,
			"size": "2",
			"status": 202,
			"total_price": 5950,
			"track_number": "WBILMTESTTRACK15"
		}
	],
	"locale": "es",
	"oof_shard": "0",
	"order_uid": "b563feb7b2b84b6test15",
	"payment": {
		"amount": 6746,
		"bank": "sberbank",
		"currency": "EUR",
		"custom_fee": 0,
		"delivery_cost": 167,
		"goods_total": 6246,
		"payment_dt": 1756117235,
		"provider": "wbpay",
		"request_id": "",
		"transaction": "b563feb7b2b84b6test15"
	},
	"shardkey": "0",
	"sm_id": 78,
	"track_number": "WBILMTESTTRACK15"
}
```

- Ответ с кодом 404:

```json
{
	"status": 404,
	"text": "order not found"
}
```

### Настройка окружения

```.env
# App settings
LOG_MODE="dev" # prod
SERVER_PORT="your_server_port"

# PostgreSQL settings
POSTGRES_USER="your_user_name"
POSTGRES_PASSWORD="yuor_password"
POSTGRES_DB="your_db_name"
POSTGRES_PORT="your_postgres_port"
POSTGRES_DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable"

# Redis settings
REDIS_PORT="your_redis_port"
REDIS_DSN="redis://redis:${REDIS_PORT}/0"

# Zookeeper settings
ZOOKEEPER_PORT="your_zookeeper_port"
ZOOKEEPER_TICK_TIME="your_zookeeper_tick_time"

# Kafka settings
KAFKA1_PORT="your_kafka1_port"
KAFKA2_PORT="your_kafka2_port"
KAFKA3_PORT="your_kafka3_port"
KAFKA_BOOTSTRAP_SERVERS="kafka1:2${KAFKA1_PORT},kafka2:2${KAFKA2_PORT},kafka3:2${KAFKA3_PORT}"

# Producer
KAFKA_CLIENT_ID="your_kafka_client_id"
KAFKA_ACKS="your_kafka_acks"
KAFKA_COMPRESSION_TYPE="your_kafka_compression_type"
KAFKA_RETRIES="your_kafka_retries"
KAFKA_BATCH_SIZE="your_kafka_batch_size"
KAFKA_LINGER_MS="your_kafka_linger_ms"
KAFKA_ENABLE_IDEMPOTENCE="your_kafka_enable_idempotence"
PRODUCER_PORT="your_kafka_producer_port"
TOPIC="your_kafka_topic"

# Consumer
KAFKA_CONSUMER_GROUP_ID="your_kafka_consumer_group_id"
KAFKA_AUTO_OFFSET_RESET="your_kafka_auto_offset_reset"
KAFKA_ENABLE_AUTO_COMMIT="your_kafka_auto_commit"
```

### Некоторые команды по работе с проектом

`make run` - старт приложения

`make test` - запуск тестов

`make clean` - удалить директорию

`make docker_build` - собрать контейнеры

`make docker_up` - поднять контейнеры

`docker down` - остановить контейнеры

### Тестирование

- ~60% покрытия юнит тестами;
- для тестирования использовался инструменты `gomock`, `pgxmock`, `redismock`.
