package models

import (
	"time"
)

type LocaleEnum string

const (
	LocaleEN LocaleEnum = "en"
	LocaleRU LocaleEnum = "ru"
	LocaleES LocaleEnum = "es"
	LocaleFR LocaleEnum = "fr"
	LocaleDE LocaleEnum = "de"
	LocaleIT LocaleEnum = "it"
	LocaleZH LocaleEnum = "zh"
	LocaleJA LocaleEnum = "ja"
	LocaleKO LocaleEnum = "ko"
	LocaleAR LocaleEnum = "ar"
)

type CurrencyEnum string

const (
	CurrencyUSD CurrencyEnum = "USD"
	CurrencyEUR CurrencyEnum = "EUR"
	CurrencyRUB CurrencyEnum = "RUB"
	CurrencyGBP CurrencyEnum = "GBP"
	CurrencyJPY CurrencyEnum = "JPY"
	CurrencyCNY CurrencyEnum = "CNY"
	CurrencyCAD CurrencyEnum = "CAD"
	CurrencyAUD CurrencyEnum = "AUD"
	CurrencyCHF CurrencyEnum = "CHF"
)

type Order struct {
	ID                int64      `json:"id" db:"id"`
	OrderUID          string     `json:"order_uid" db:"order_uid"`
	TrackNumber       string     `json:"track_number" db:"track_number"`
	Entry             string     `json:"entry" db:"entry"`
	Locale            LocaleEnum `json:"locale" db:"locale"`
	InternalSignature string     `json:"internal_signature" db:"internal_signature"`
	CustomerID        string     `json:"customer_id" db:"customer_id"`
	DeliveryService   string     `json:"delivery_service" db:"delivery_service"`
	Shardkey          string     `json:"shardkey" db:"shardkey"`
	SmID              int        `json:"sm_id" db:"sm_id"`
	OofShard          string     `json:"oof_shard" db:"oof_shard"`
	DateCreated       time.Time  `json:"date_created" db:"date_created"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`

	Delivery *Delivery `json:"delivery,omitempty" db:"-"`
	Payment  *Payment  `json:"payment,omitempty" db:"-"`
	Items    []Item    `json:"items,omitempty" db:"-"`
}

type Delivery struct {
	ID        int64     `json:"id" db:"id"`
	OrderID   int64     `json:"order_id" db:"order_id"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	Zip       string    `json:"zip" db:"zip"`
	City      string    `json:"city" db:"city"`
	Address   string    `json:"address" db:"address"`
	Region    string    `json:"region" db:"region"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Payment struct {
	ID           int64        `json:"id" db:"id"`
	OrderID      int64        `json:"order_id" db:"order_id"`
	Transaction  string       `json:"transaction" db:"transaction"`
	RequestID    string       `json:"request_id" db:"request_id"`
	Currency     CurrencyEnum `json:"currency" db:"currency"`
	Provider     string       `json:"provider" db:"provider"`
	Amount       int          `json:"amount" db:"amount"`
	PaymentDt    int          `json:"payment_dt" db:"payment_dt"`
	Bank         string       `json:"bank" db:"bank"`
	DeliveryCost int          `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   int          `json:"goods_total" db:"goods_total"`
	CustomFee    int          `json:"custom_fee" db:"custom_fee"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

type Item struct {
	ID          int64     `json:"id" db:"id"`
	OrderID     int64     `json:"order_id" db:"order_id"`
	ChrtID      int       `json:"chrt_id" db:"chrt_id"`
	TrackNumber string    `json:"track_number" db:"track_number"`
	Price       int       `json:"price" db:"price"`
	Rid         string    `json:"rid" db:"rid"`
	Name        string    `json:"name" db:"name"`
	Sale        int       `json:"sale" db:"sale"`
	Size        string    `json:"size" db:"size"`
	TotalPrice  int       `json:"total_price" db:"total_price"`
	NmID        int       `json:"nm_id" db:"nm_id"`
	Brand       string    `json:"brand" db:"brand"`
	Status      int       `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type OrderRequest struct {
	OrderUID          string          `json:"order_uid"`
	TrackNumber       string          `json:"track_number"`
	Entry             string          `json:"entry"`
	Locale            LocaleEnum      `json:"locale"`
	InternalSignature string          `json:"internal_signature"`
	CustomerID        string          `json:"customer_id"`
	DeliveryService   string          `json:"delivery_service"`
	Shardkey          string          `json:"shardkey"`
	SmID              int             `json:"sm_id"`
	DateCreated       time.Time       `json:"date_created"`
	OofShard          string          `json:"oof_shard"`
	Delivery          DeliveryRequest `json:"delivery"`
	Payment           PaymentRequest  `json:"payment"`
	Items             []ItemRequest   `json:"items"`
}

type DeliveryRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type PaymentRequest struct {
	Transaction  string       `json:"transaction"`
	RequestID    string       `json:"request_id"`
	Currency     CurrencyEnum `json:"currency"`
	Provider     string       `json:"provider"`
	Amount       int          `json:"amount"`
	PaymentDt    int          `json:"payment_dt"`
	Bank         string       `json:"bank"`
	DeliveryCost int          `json:"delivery_cost"`
	GoodsTotal   int          `json:"goods_total"`
	CustomFee    int          `json:"custom_fee"`
}

type ItemRequest struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}
