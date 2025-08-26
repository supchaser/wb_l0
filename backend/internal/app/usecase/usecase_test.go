package usecase

import (
	"context"
	"errors"
	"testing"

	gomock "github.com/golang/mock/gomock"
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

func TestAppUsecase_GetOrderByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		inputOrderUID string
		mockSetup     func(*mock_app.MockAppRepository)
		expectedOrder *models.Order
		expectedError error
		errorContains string
	}{
		{
			name:          "Success",
			inputOrderUID: "valid-order-uid-123",
			mockSetup: func(mockRepo *mock_app.MockAppRepository) {
				mockRepo.EXPECT().
					GetOrderByID(gomock.Any(), "valid-order-uid-123").
					Return(&models.Order{
						OrderUID:    "valid-order-uid-123",
						TrackNumber: "WBILMTESTTRACK",
						Entry:       "WBIL",
					}, nil)
			},
			expectedOrder: &models.Order{
				OrderUID:    "valid-order-uid-123",
				TrackNumber: "WBILMTESTTRACK",
				Entry:       "WBIL",
			},
			expectedError: nil,
		},
		{
			name:          "InvalidOrderUID",
			inputOrderUID: "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@",
			mockSetup: func(mockRepo *mock_app.MockAppRepository) {
				mockRepo.EXPECT().
					GetOrderByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedOrder: nil,
			expectedError: errs.ErrValidation,
		},
		{
			name:          "NotFound",
			inputOrderUID: "non-existent-order-456",
			mockSetup: func(mockRepo *mock_app.MockAppRepository) {
				mockRepo.EXPECT().
					GetOrderByID(gomock.Any(), "non-existent-order-456").
					Return(nil, errs.ErrNotFound)
			},
			expectedOrder: nil,
			expectedError: errs.ErrNotFound,
		},
		{
			name:          "RepositoryError",
			inputOrderUID: "valid-order-789",
			mockSetup: func(mockRepo *mock_app.MockAppRepository) {
				mockRepo.EXPECT().
					GetOrderByID(gomock.Any(), "valid-order-789").
					Return(nil, errors.New("database connection failed"))
			},
			expectedOrder: nil,
			expectedError: errors.New("Usecase.GetOrderByID: failed to get order:"),
			errorContains: "Usecase.GetOrderByID: failed to get order:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockAppRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			uc := &AppUsecase{
				orderRepository: mockRepo,
			}
			result, err := uc.GetOrderByID(context.Background(), tt.inputOrderUID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				} else {
					assert.ErrorIs(t, err, tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedOrder, result)
		})
	}
}

func TestAppUsecase_GetOrderByID_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCases := []struct {
		name          string
		orderUID      string
		expectedError error
	}{
		{
			name:          "EmptyOrderUID",
			orderUID:      "",
			expectedError: errs.ErrValidation,
		},
		{
			name:          "TooLongOrderUID",
			orderUID:      "https://music.yandex.ru/track/141708966?utm_source=desktop&utm_medium=copy_link https://music.yandex.ru/track/141708966?utm_source=desktop&utm_medium=copy_linkhttps://music.yandex.ru/track/141708966?utm_source=desktop&utm_medium=copy_link",
			expectedError: errs.ErrValidation,
		},
		{
			name:          "InvalidFormatOrderUID",
			orderUID:      "invalid@uid#",
			expectedError: errs.ErrValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mock_app.NewMockAppRepository(ctrl)
			mockRepo.EXPECT().
				GetOrderByID(gomock.Any(), gomock.Any()).
				Times(0)

			uc := &AppUsecase{
				orderRepository: mockRepo,
			}
			result, err := uc.GetOrderByID(context.Background(), tc.orderUID)

			assert.Error(t, err)
			assert.ErrorIs(t, err, tc.expectedError)
			assert.Nil(t, result)
		})
	}
}

func TestCreateAppUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_app.NewMockAppRepository(ctrl)

	uc := CreateAppUsecase(mockRepo)

	assert.NotNil(t, uc)
	assert.IsType(t, &AppUsecase{}, uc)
}
