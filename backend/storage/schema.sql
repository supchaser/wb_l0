CREATE TYPE locale_enum AS ENUM (
    'en',
    'ru',
    'es',
    'fr',
    'de',
    'it',
    'zh',
    'ja',
    'ko',
    'ar'
);

CREATE TYPE currency_enum AS ENUM (
    'USD',
    'EUR',
    'RUB',
    'GBP',
    'JPY',
    'CNY',
    'CAD',
    'AUD',
    'CHF',
);

CREATE TABLE
    IF NOT EXISTS "order" (
        id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        order_uid TEXT NOT NULL UNIQUE,
        track_number TEXT NOT NULL,
        entry TEXT NOT NULL,
        locale locale_enum NOT NULL DEFAULT 'ru',
        internal_signature TEXT DEFAULT '',
        customer_id TEXT NOT NULL,
        delivery_service TEXT NOT NULL,
        shardkey TEXT NOT NULL,
        sm_id INTEGER NOT NULL,
        oof_shard TEXT NOT NULL,
        date_created TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    IF NOT EXISTS delivery (
        id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        order_id BIGINT NOT NULL UNIQUE,
        name TEXT NOT NULL,
        phone TEXT NOT NULL,
        zip TEXT NOT NULL,
        city TEXT NOT NULL,
        address TEXT NOT NULL,
        region TEXT NOT NULL,
        email TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (order_id) REFERENCES "order" (id) ON UPDATE CASCADE ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS payment (
        id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        order_id BIGINT NOT NULL,
        transaction TEXT NOT NULL UNIQUE,
        request_id TEXT DEFAULT '',
        currency currency_enum NOT NULL DEFAULT 'USD',
        provider TEXT NOT NULL DEFAULT 'wbpay',
        amount INTEGER NOT NULL,
        payment_dt INTEGER NOT NULL,
        bank TEXT NOT NULL,
        delivery_cost INTEGER NOT NULL,
        goods_total INTEGER NOT NULL,
        custom_fee INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (order_id) REFERENCES "order" (id) ON UPDATE CASCADE ON DELETE CASCADE
    );

CREATE TABLE
    IF NOT EXISTS item (
        id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        order_id BIGINT NOT NULL,
        chrt_id INTEGER NOT NULL,
        track_number TEXT NOT NULL,
        price INTEGER NOT NULL,
        rid TEXT NOT NULL,
        name TEXT NOT NULL,
        sale INTEGER NOT NULL,
        size TEXT NOT NULL,
        total_price INTEGER NOT NULL,
        nm_id INTEGER NOT NULL,
        brand TEXT NOT NULL,
        status INTEGER NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (order_id) REFERENCES "order" (id) ON UPDATE CASCADE ON DELETE CASCADE
    );

CREATE UNIQUE INDEX idx_order_order_uid ON "order" (order_uid);

CREATE INDEX idx_order_track_number ON "order" (track_number);

CREATE INDEX idx_order_customer_id ON "order" (customer_id);

CREATE INDEX idx_order_date_created ON "order" (date_created);

CREATE INDEX idx_order_sm_id ON "order" (sm_id);

CREATE INDEX idx_delivery_order_id ON delivery (order_id);

CREATE INDEX idx_delivery_email ON delivery (email);

CREATE INDEX idx_payment_order_id ON payment (order_id);

CREATE INDEX idx_payment_transaction ON payment (transaction);

CREATE INDEX idx_item_order_id ON item (order_id);

CREATE INDEX idx_item_chrt_id ON item (chrt_id);

CREATE INDEX idx_item_nm_id ON item (nm_id);
