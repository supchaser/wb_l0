package validate

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
)

func TestValidateOrderRequest(t *testing.T) {
	tests := []struct {
		name    string
		order   *models.OrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidOrder",
			order:   createValidOrderRequest(),
			wantErr: false,
		},
		{
			name: "InvalidOrderUID",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.OrderUID = "invalid@uid#"
				return order
			}(),
			wantErr: true,
			errMsg:  "order_uid contains invalid characters",
		},
		{
			name: "EmptyOrderUID",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.OrderUID = ""
				return order
			}(),
			wantErr: true,
			errMsg:  "order_uid is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderRequest(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMainOrder(t *testing.T) {
	tests := []struct {
		name    string
		order   *models.OrderRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidMainOrder",
			order:   createValidOrderRequest(),
			wantErr: false,
		},
		{
			name: "InvalidTrackNumber",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.TrackNumber = "invalid@track"
				return order
			}(),
			wantErr: true,
			errMsg:  "track_number can only contain uppercase letters and numbers",
		},
		{
			name: "InvalidLocale",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.Locale = "invalid"
				return order
			}(),
			wantErr: true,
			errMsg:  "invalid locale value",
		},
		{
			name: "FutureDateCreated",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.DateCreated = time.Now().Add(48 * time.Hour)
				return order
			}(),
			wantErr: true,
			errMsg:  "date_created cannot be in the future",
		},
		{
			name: "ZeroDateCreated",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.DateCreated = time.Time{}
				return order
			}(),
			wantErr: true,
			errMsg:  "date_created is required",
		},
		{
			name: "InvalidShardkey",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.Shardkey = "abc"
				return order
			}(),
			wantErr: true,
			errMsg:  "shardkey can only contain numbers",
		},
		{
			name: "InvalidOofShard",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.OofShard = "abc"
				return order
			}(),
			wantErr: true,
			errMsg:  "oof_shard can only contain numbers",
		},
		{
			name: "NegativeSmID",
			order: func() *models.OrderRequest {
				order := createValidOrderRequest()
				order.SmID = -1
				return order
			}(),
			wantErr: true,
			errMsg:  "sm_id must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMainOrder(tt.order)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, errors.Is(err, errs.ErrValidation))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDelivery(t *testing.T) {
	tests := []struct {
		name     string
		delivery *models.DeliveryRequest
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "ValidDelivery",
			delivery: createValidDeliveryRequest(),
			wantErr:  false,
		},
		{
			name: "InvalidPhone",
			delivery: func() *models.DeliveryRequest {
				delivery := createValidDeliveryRequest()
				delivery.Phone = "invalid_phone"
				return delivery
			}(),
			wantErr: true,
			errMsg:  "delivery phone contains invalid characters",
		},
		{
			name: "InvalidEmail",
			delivery: func() *models.DeliveryRequest {
				delivery := createValidDeliveryRequest()
				delivery.Email = "invalid-email"
				return delivery
			}(),
			wantErr: true,
			errMsg:  "delivery email is invalid",
		},
		{
			name: "InvalidZip",
			delivery: func() *models.DeliveryRequest {
				delivery := createValidDeliveryRequest()
				delivery.Zip = "invalid@zip#"
				return delivery
			}(),
			wantErr: true,
			errMsg:  "delivery zip contains invalid characters",
		},
		{
			name: "EmptyName",
			delivery: func() *models.DeliveryRequest {
				delivery := createValidDeliveryRequest()
				delivery.Name = ""
				return delivery
			}(),
			wantErr: true,
			errMsg:  "delivery name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDelivery(tt.delivery)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, errors.Is(err, errs.ErrValidation))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePayment(t *testing.T) {
	tests := []struct {
		name    string
		payment *models.PaymentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidPayment",
			payment: createValidPaymentRequest(),
			wantErr: false,
		},
		{
			name: "InvalidTransaction",
			payment: func() *models.PaymentRequest {
				payment := createValidPaymentRequest()
				payment.Transaction = "invalid@trans#"
				return payment
			}(),
			wantErr: true,
			errMsg:  "payment transaction contains invalid characters",
		},
		{
			name: "NegativeAmount",
			payment: func() *models.PaymentRequest {
				payment := createValidPaymentRequest()
				payment.Amount = -100
				return payment
			}(),
			wantErr: true,
			errMsg:  "payment amount cannot be negative",
		},
		{
			name: "InvalidCurrency",
			payment: func() *models.PaymentRequest {
				payment := createValidPaymentRequest()
				payment.Currency = "INVALID"
				return payment
			}(),
			wantErr: true,
			errMsg:  "invalid payment currency",
		},
		{
			name: "NegativePaymentDt",
			payment: func() *models.PaymentRequest {
				payment := createValidPaymentRequest()
				payment.PaymentDt = 0
				return payment
			}(),
			wantErr: true,
			errMsg:  "payment_dt must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePayment(tt.payment)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, errors.Is(err, errs.ErrValidation))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateItems(t *testing.T) {
	tests := []struct {
		name    string
		items   []models.ItemRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidItems",
			items:   createValidItemsRequest(),
			wantErr: false,
		},
		{
			name:    "EmptyItems",
			items:   []models.ItemRequest{},
			wantErr: true,
			errMsg:  "at least one item is required",
		},
		{
			name: "InvalidItemRid",
			items: func() []models.ItemRequest {
				items := createValidItemsRequest()
				items[0].Rid = "invalid@rid#"
				return items
			}(),
			wantErr: true,
			errMsg:  "item[0].rid contains invalid characters",
		},
		{
			name: "NegativeItemPrice",
			items: func() []models.ItemRequest {
				items := createValidItemsRequest()
				items[0].Price = -100
				return items
			}(),
			wantErr: true,
			errMsg:  "item[0].price must be positive",
		},
		{
			name: "NegativeItemStatus",
			items: func() []models.ItemRequest {
				items := createValidItemsRequest()
				items[0].Status = -1
				return items
			}(),
			wantErr: true,
			errMsg:  "item[0].status cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateItems(tt.items)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, errors.Is(err, errs.ErrValidation))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOrderUID(t *testing.T) {
	tests := []struct {
		name     string
		orderUID string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "ValidOrderUID",
			orderUID: "test123-abc_456",
			wantErr:  false,
		},
		{
			name:     "EmptyOrderUID",
			orderUID: "",
			wantErr:  true,
			errMsg:   "order UID cannot be empty",
		},
		{
			name:     "TooLongOrderUID",
			orderUID: "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz", // 52 characters
			wantErr:  true,
			errMsg:   "order UID too long",
		},
		{
			name:     "InvalidCharacters",
			orderUID: "test@123#",
			wantErr:  true,
			errMsg:   "order_uid contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrderUID(tt.orderUID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidLocale(t *testing.T) {
	tests := []struct {
		locale models.LocaleEnum
		want   bool
	}{
		{models.LocaleEN, true},
		{models.LocaleRU, true},
		{models.LocaleES, true},
		{models.LocaleFR, true},
		{models.LocaleDE, true},
		{models.LocaleIT, true},
		{models.LocaleZH, true},
		{models.LocaleJA, true},
		{models.LocaleKO, true},
		{models.LocaleAR, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.locale), func(t *testing.T) {
			result := isValidLocale(tt.locale)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsValidCurrency(t *testing.T) {
	tests := []struct {
		currency models.CurrencyEnum
		want     bool
	}{
		{models.CurrencyUSD, true},
		{models.CurrencyEUR, true},
		{models.CurrencyRUB, true},
		{models.CurrencyGBP, true},
		{models.CurrencyJPY, true},
		{models.CurrencyCNY, true},
		{models.CurrencyCAD, true},
		{models.CurrencyAUD, true},
		{models.CurrencyCHF, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.currency), func(t *testing.T) {
			result := isValidCurrency(tt.currency)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMaxLengthValidations(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*models.OrderRequest)
		errMsg string
	}{
		{
			name: "LongOrderUID",
			setup: func(o *models.OrderRequest) {
				o.OrderUID = string(make([]rune, MaxOrderUIDLength+1))
			},
			errMsg: fmt.Sprintf("order_uid cannot be longer than %d characters", MaxOrderUIDLength),
		},
		{
			name: "LongTrackNumber",
			setup: func(o *models.OrderRequest) {
				o.TrackNumber = string(make([]rune, MaxTrackNumberLength+1))
			},
			errMsg: fmt.Sprintf("track_number cannot be longer than %d characters", MaxTrackNumberLength),
		},
		{
			name: "LongInternalSignature",
			setup: func(o *models.OrderRequest) {
				o.InternalSignature = string(make([]rune, MaxInternalSigLength+1))
			},
			errMsg: fmt.Sprintf("internal_signature cannot be longer than %d characters", MaxInternalSigLength),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createValidOrderRequest()
			tt.setup(order)

			err := ValidateMainOrder(order)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func createValidOrderRequest() *models.OrderRequest {
	return &models.OrderRequest{
		OrderUID:          "test123-abc_456",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            models.LocaleEN,
		InternalSignature: "",
		CustomerID:        "test_customer",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now().Add(-24 * time.Hour),
		OofShard:          "1",
		Delivery:          *createValidDeliveryRequest(),
		Payment:           *createValidPaymentRequest(),
		Items:             createValidItemsRequest(),
	}
}

func createValidDeliveryRequest() *models.DeliveryRequest {
	return &models.DeliveryRequest{
		Name:    "Test Testov",
		Phone:   "+9720000000",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	}
}

func createValidPaymentRequest() *models.PaymentRequest {
	return &models.PaymentRequest{
		Transaction:  "test123-abc",
		RequestID:    "",
		Currency:     models.CurrencyUSD,
		Provider:     "wbpay",
		Amount:       1817,
		PaymentDt:    1637907727,
		Bank:         "alpha",
		DeliveryCost: 1500,
		GoodsTotal:   317,
		CustomFee:    0,
	}
}

func createValidItemsRequest() []models.ItemRequest {
	return []models.ItemRequest{
		{
			ChrtID:      9934930,
			TrackNumber: "WBILMTESTTRACK",
			Price:       453,
			Rid:         "ab4219087a764ae0btest",
			Name:        "Mascaras",
			Sale:        30,
			Size:        "0",
			TotalPrice:  317,
			NmID:        2389212,
			Brand:       "Vivienne Sabo",
			Status:      202,
		},
	}
}
