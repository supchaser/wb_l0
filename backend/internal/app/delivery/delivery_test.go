package delivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	mock_app "github.com/supchaser/wb_l0/internal/app/mocks"
	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
	"github.com/supchaser/wb_l0/internal/utils/logger"
)

func TestMain(m *testing.M) {
	logger.InitTestLogger()
	m.Run()
}

func TestAppDelivery_GetOrderByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	testTime := time.Now()
	testOrder := &models.Order{
		OrderUID:          "test123",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       testTime,
		OofShard:          "1",
		Delivery: &models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: &models.Payment{
			Transaction:  "test123",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
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
		},
	}

	tests := []struct {
		name           string
		orderUID       string
		mockSetup      func()
		expectedStatus int
		validateFunc   func(t *testing.T, body []byte)
	}{
		{
			name:     "Success",
			orderUID: "test123",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetOrderByID(gomock.Any(), "test123").
					Return(testOrder, nil)
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]any
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, "test123", response["order_uid"])
				assert.Equal(t, "WBILMTESTTRACK", response["track_number"])
				assert.Equal(t, testTime.Format(time.RFC3339), response["date_created"])

				delivery := response["delivery"].(map[string]any)
				assert.Equal(t, "Test Testov", delivery["name"])
				assert.Equal(t, "+9720000000", delivery["phone"])

				payment := response["payment"].(map[string]any)
				assert.Equal(t, "test123", payment["transaction"])
				assert.Equal(t, float64(1817), payment["amount"])

				items := response["items"].([]any)
				assert.Len(t, items, 1)
				item := items[0].(map[string]any)
				assert.Equal(t, float64(9934930), item["chrt_id"])
				assert.Equal(t, "Mascaras", item["name"])
			},
		},
		{
			name:     "OrderNotFound",
			orderUID: "nonexistent",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetOrderByID(gomock.Any(), "nonexistent").
					Return(nil, errs.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]any
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(404), response["status"])
				assert.Equal(t, "order not found", response["text"])
			},
		},
		{
			name:     "InternalServerError",
			orderUID: "test123",
			mockSetup: func() {
				mockUsecase.EXPECT().
					GetOrderByID(gomock.Any(), "test123").
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]any
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(500), response["status"])
				assert.Equal(t, "internal server error", response["text"])
			},
		},
		{
			name:     "MissingOrderUID",
			orderUID: "",
			mockSetup: func() {
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]any
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(400), response["status"])
				assert.Equal(t, "order_uid is required", response["text"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/orders/"+tt.orderUID, nil)
			if tt.orderUID != "" {
				req = mux.SetURLVars(req, map[string]string{"order_uid": tt.orderUID})
			}

			w := httptest.NewRecorder()

			appDelivery.GetOrderByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateFunc(t, w.Body.Bytes())
		})
	}
}

func TestAppDelivery_convertToResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	testTime := time.Now()
	order := &models.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		DateCreated: testTime,
		Delivery: &models.Delivery{
			Name:  "Test Testov",
			Phone: "+9720000000",
		},
		Payment: &models.Payment{
			Transaction: "test123",
			Amount:      1817,
		},
		Items: []models.Item{
			{
				ChrtID: 9934930,
				Name:   "Mascaras",
			},
		},
	}

	response := appDelivery.convertToResponse(order)

	assert.Equal(t, "test123", response["order_uid"])
	assert.Equal(t, "WBILMTESTTRACK", response["track_number"])
	assert.Equal(t, testTime.Format(time.RFC3339), response["date_created"])

	delivery := response["delivery"].(map[string]any)
	assert.Equal(t, "Test Testov", delivery["name"])
	assert.Equal(t, "+9720000000", delivery["phone"])

	payment := response["payment"].(map[string]any)
	assert.Equal(t, "test123", payment["transaction"])
	assert.Equal(t, int(1817), payment["amount"])

	items := response["items"].([]map[string]any)
	assert.Len(t, items, 1)
	assert.Equal(t, int(9934930), items[0]["chrt_id"])
	assert.Equal(t, "Mascaras", items[0]["name"])
}

func TestAppDelivery_convertToResponse_NilFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	order := &models.Order{
		OrderUID:    "test123",
		TrackNumber: "WBILMTESTTRACK",
		DateCreated: time.Now(),
		Delivery:    nil,
		Payment:     nil,
		Items:       nil,
	}

	response := appDelivery.convertToResponse(order)

	assert.Nil(t, response["delivery"])
	assert.Nil(t, response["payment"])
	assert.Nil(t, response["items"])
}

func TestAppDelivery_convertDeliveryToResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	delivery := &models.Delivery{
		Name:    "Test Testov",
		Phone:   "+9720000000",
		Zip:     "2639809",
		City:    "Kiryat Mozkin",
		Address: "Ploshad Mira 15",
		Region:  "Kraiot",
		Email:   "test@gmail.com",
	}

	response := appDelivery.convertDeliveryToResponse(delivery)

	assert.Equal(t, "Test Testov", response["name"])
	assert.Equal(t, "+9720000000", response["phone"])
	assert.Equal(t, "2639809", response["zip"])
	assert.Equal(t, "Kiryat Mozkin", response["city"])
}

func TestAppDelivery_convertDeliveryToResponse_Nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	response := appDelivery.convertDeliveryToResponse(nil)
	assert.Nil(t, response)
}

func TestAppDelivery_convertPaymentToResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	payment := &models.Payment{
		Transaction:  "test123",
		RequestID:    "",
		Currency:     "USD",
		Provider:     "wbpay",
		Amount:       1817,
		PaymentDt:    1637907727,
		Bank:         "alpha",
		DeliveryCost: 1500,
		GoodsTotal:   317,
		CustomFee:    0,
	}

	response := appDelivery.convertPaymentToResponse(payment)

	assert.Equal(t, "test123", response["transaction"])
	assert.Equal(t, int(1817), response["amount"])
	assert.Equal(t, int(1500), response["delivery_cost"])
	assert.Equal(t, int(317), response["goods_total"])
}

func TestAppDelivery_convertPaymentToResponse_Nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	response := appDelivery.convertPaymentToResponse(nil)
	assert.Nil(t, response)
}

func TestAppDelivery_convertItemsToResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	items := []models.Item{
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
		{
			ChrtID:      9934931,
			TrackNumber: "WBILMTESTTRACK2",
			Price:       500,
			Rid:         "ab4219087a764ae0btest2",
			Name:        "Lipstick",
			Sale:        20,
			Size:        "1",
			TotalPrice:  400,
			NmID:        2389213,
			Brand:       "Maybelline",
			Status:      203,
		},
	}

	response := appDelivery.convertItemsToResponse(items)

	assert.Len(t, response, 2)
	assert.Equal(t, int(9934930), response[0]["chrt_id"])
	assert.Equal(t, "Mascaras", response[0]["name"])
	assert.Equal(t, int(9934931), response[1]["chrt_id"])
	assert.Equal(t, "Lipstick", response[1]["name"])
}

func TestAppDelivery_convertItemsToResponse_Nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	response := appDelivery.convertItemsToResponse(nil)
	assert.Nil(t, response)
}

func TestAppDelivery_convertItemsToResponse_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock_app.NewMockAppUsecase(ctrl)
	appDelivery := CreateAppDelivery(mockUsecase)

	response := appDelivery.convertItemsToResponse([]models.Item{})
	assert.Empty(t, response)
}
