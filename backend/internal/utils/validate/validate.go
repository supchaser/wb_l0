package validate

import (
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/supchaser/wb_l0/internal/app/models"
	"github.com/supchaser/wb_l0/internal/utils/errs"
)

const (
	MaxOrderUIDLength       = 50
	MaxTrackNumberLength    = 50
	MaxEntryLength          = 10
	MaxInternalSigLength    = 100
	MaxCustomerIDLength     = 50
	MaxDeliveryServiceLen   = 50
	MaxShardkeyLength       = 10
	MaxOofShardLength       = 10
	MaxDeliveryNameLength   = 100
	MaxDeliveryPhoneLength  = 20
	MaxDeliveryZipLength    = 20
	MaxDeliveryCityLength   = 100
	MaxDeliveryAddrLength   = 200
	MaxDeliveryRegionLength = 100
	MaxDeliveryEmailLength  = 255
	MaxPaymentTransLength   = 50
	MaxPaymentReqIDLength   = 50
	MaxPaymentProviderLen   = 50
	MaxPaymentBankLength    = 50
	MaxItemTrackNumberLen   = 50
	MaxItemRidLength        = 50
	MaxItemNameLength       = 200
	MaxItemSizeLength       = 10
	MaxItemBrandLength      = 100
)

var (
	orderUIDRegex     = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	trackNumberRegex  = regexp.MustCompile(`^[A-Z0-9]+$`)
	entryRegex        = regexp.MustCompile(`^[A-Z]+$`)
	shardkeyRegex     = regexp.MustCompile(`^[0-9]+$`)
	oofShardRegex     = regexp.MustCompile(`^[0-9]+$`)
	phoneRegex        = regexp.MustCompile(`^\+?[0-9\s\-\(\)]+$`)
	zipRegex          = regexp.MustCompile(`^[0-9A-Za-z\-]+$`)
	emailRegex        = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	paymentTransRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	itemRidRegex      = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

func ValidateOrderRequest(order *models.OrderRequest) error {
	if err := ValidateMainOrder(order); err != nil {
		return err
	}

	if err := ValidateDelivery(&order.Delivery); err != nil {
		return fmt.Errorf("delivery validation failed: %w", err)
	}

	if err := ValidatePayment(&order.Payment); err != nil {
		return fmt.Errorf("payment validation failed: %w", err)
	}

	if err := ValidateItems(order.Items); err != nil {
		return fmt.Errorf("items validation failed: %w", err)
	}

	return nil
}

func ValidateMainOrder(order *models.OrderRequest) error {
	if order.OrderUID == "" {
		return fmt.Errorf("%w: order_uid is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.OrderUID) > MaxOrderUIDLength {
		return fmt.Errorf("%w: order_uid cannot be longer than %d characters", errs.ErrValidation, MaxOrderUIDLength)
	}
	if !orderUIDRegex.MatchString(order.OrderUID) {
		return fmt.Errorf("%w: order_uid contains invalid characters", errs.ErrValidation)
	}

	if order.TrackNumber == "" {
		return fmt.Errorf("%w: track_number is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.TrackNumber) > MaxTrackNumberLength {
		return fmt.Errorf("%w: track_number cannot be longer than %d characters", errs.ErrValidation, MaxTrackNumberLength)
	}
	if !trackNumberRegex.MatchString(order.TrackNumber) {
		return fmt.Errorf("%w: track_number can only contain uppercase letters and numbers", errs.ErrValidation)
	}

	if order.Entry == "" {
		return fmt.Errorf("%w: entry is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.Entry) > MaxEntryLength {
		return fmt.Errorf("%w: entry cannot be longer than %d characters", errs.ErrValidation, MaxEntryLength)
	}
	if !entryRegex.MatchString(order.Entry) {
		return fmt.Errorf("%w: entry can only contain uppercase letters", errs.ErrValidation)
	}

	if order.Locale == "" {
		return fmt.Errorf("%w: locale is required", errs.ErrValidation)
	}
	if !isValidLocale(order.Locale) {
		return fmt.Errorf("%w: invalid locale value", errs.ErrValidation)
	}

	if utf8.RuneCountInString(order.InternalSignature) > MaxInternalSigLength {
		return fmt.Errorf("%w: internal_signature cannot be longer than %d characters", errs.ErrValidation, MaxInternalSigLength)
	}

	if order.CustomerID == "" {
		return fmt.Errorf("%w: customer_id is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.CustomerID) > MaxCustomerIDLength {
		return fmt.Errorf("%w: customer_id cannot be longer than %d characters", errs.ErrValidation, MaxCustomerIDLength)
	}

	if order.DeliveryService == "" {
		return fmt.Errorf("%w: delivery_service is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.DeliveryService) > MaxDeliveryServiceLen {
		return fmt.Errorf("%w: delivery_service cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryServiceLen)
	}

	if order.Shardkey == "" {
		return fmt.Errorf("%w: shardkey is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.Shardkey) > MaxShardkeyLength {
		return fmt.Errorf("%w: shardkey cannot be longer than %d characters", errs.ErrValidation, MaxShardkeyLength)
	}
	if !shardkeyRegex.MatchString(order.Shardkey) {
		return fmt.Errorf("%w: shardkey can only contain numbers", errs.ErrValidation)
	}

	if order.SmID <= 0 {
		return fmt.Errorf("%w: sm_id must be positive", errs.ErrValidation)
	}

	if order.OofShard == "" {
		return fmt.Errorf("%w: oof_shard is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(order.OofShard) > MaxOofShardLength {
		return fmt.Errorf("%w: oof_shard cannot be longer than %d characters", errs.ErrValidation, MaxOofShardLength)
	}
	if !oofShardRegex.MatchString(order.OofShard) {
		return fmt.Errorf("%w: oof_shard can only contain numbers", errs.ErrValidation)
	}

	if order.DateCreated.IsZero() {
		return fmt.Errorf("%w: date_created is required", errs.ErrValidation)
	}
	if order.DateCreated.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("%w: date_created cannot be in the future", errs.ErrValidation)
	}

	return nil
}

func ValidateDelivery(delivery *models.DeliveryRequest) error {
	if delivery.Name == "" {
		return fmt.Errorf("%w: delivery name is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Name) > MaxDeliveryNameLength {
		return fmt.Errorf("%w: delivery name cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryNameLength)
	}

	if delivery.Phone == "" {
		return fmt.Errorf("%w: delivery phone is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Phone) > MaxDeliveryPhoneLength {
		return fmt.Errorf("%w: delivery phone cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryPhoneLength)
	}
	if !phoneRegex.MatchString(delivery.Phone) {
		return fmt.Errorf("%w: delivery phone contains invalid characters", errs.ErrValidation)
	}

	if delivery.Zip == "" {
		return fmt.Errorf("%w: delivery zip is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Zip) > MaxDeliveryZipLength {
		return fmt.Errorf("%w: delivery zip cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryZipLength)
	}
	if !zipRegex.MatchString(delivery.Zip) {
		return fmt.Errorf("%w: delivery zip contains invalid characters", errs.ErrValidation)
	}

	if delivery.City == "" {
		return fmt.Errorf("%w: delivery city is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.City) > MaxDeliveryCityLength {
		return fmt.Errorf("%w: delivery city cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryCityLength)
	}

	if delivery.Address == "" {
		return fmt.Errorf("%w: delivery address is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Address) > MaxDeliveryAddrLength {
		return fmt.Errorf("%w: delivery address cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryAddrLength)
	}

	if delivery.Region == "" {
		return fmt.Errorf("%w: delivery region is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Region) > MaxDeliveryRegionLength {
		return fmt.Errorf("%w: delivery region cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryRegionLength)
	}

	if delivery.Email == "" {
		return fmt.Errorf("%w: delivery email is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(delivery.Email) > MaxDeliveryEmailLength {
		return fmt.Errorf("%w: delivery email cannot be longer than %d characters", errs.ErrValidation, MaxDeliveryEmailLength)
	}
	if !emailRegex.MatchString(delivery.Email) {
		return fmt.Errorf("%w: delivery email is invalid", errs.ErrValidation)
	}

	return nil
}

func ValidatePayment(payment *models.PaymentRequest) error {
	if payment.Transaction == "" {
		return fmt.Errorf("%w: payment transaction is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(payment.Transaction) > MaxPaymentTransLength {
		return fmt.Errorf("%w: payment transaction cannot be longer than %d characters", errs.ErrValidation, MaxPaymentTransLength)
	}
	if !paymentTransRegex.MatchString(payment.Transaction) {
		return fmt.Errorf("%w: payment transaction contains invalid characters", errs.ErrValidation)
	}

	if utf8.RuneCountInString(payment.RequestID) > MaxPaymentReqIDLength {
		return fmt.Errorf("%w: payment request_id cannot be longer than %d characters", errs.ErrValidation, MaxPaymentReqIDLength)
	}

	if payment.Currency == "" {
		return fmt.Errorf("%w: payment currency is required", errs.ErrValidation)
	}
	if !isValidCurrency(payment.Currency) {
		return fmt.Errorf("%w: invalid payment currency", errs.ErrValidation)
	}

	if payment.Provider == "" {
		return fmt.Errorf("%w: payment provider is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(payment.Provider) > MaxPaymentProviderLen {
		return fmt.Errorf("%w: payment provider cannot be longer than %d characters", errs.ErrValidation, MaxPaymentProviderLen)
	}

	if payment.Amount < 0 {
		return fmt.Errorf("%w: payment amount cannot be negative", errs.ErrValidation)
	}

	if payment.PaymentDt <= 0 {
		return fmt.Errorf("%w: payment_dt must be positive", errs.ErrValidation)
	}

	if payment.Bank == "" {
		return fmt.Errorf("%w: payment bank is required", errs.ErrValidation)
	}
	if utf8.RuneCountInString(payment.Bank) > MaxPaymentBankLength {
		return fmt.Errorf("%w: payment bank cannot be longer than %d characters", errs.ErrValidation, MaxPaymentBankLength)
	}

	if payment.DeliveryCost < 0 {
		return fmt.Errorf("%w: delivery_cost cannot be negative", errs.ErrValidation)
	}

	if payment.GoodsTotal < 0 {
		return fmt.Errorf("%w: goods_total cannot be negative", errs.ErrValidation)
	}

	if payment.CustomFee < 0 {
		return fmt.Errorf("%w: custom_fee cannot be negative", errs.ErrValidation)
	}

	return nil
}

func ValidateItems(items []models.ItemRequest) error {
	if len(items) == 0 {
		return fmt.Errorf("%w: at least one item is required", errs.ErrValidation)
	}

	for i, item := range items {
		if err := validateItem(item, i); err != nil {
			return err
		}
	}

	return nil
}

func validateItem(item models.ItemRequest, index int) error {
	if item.ChrtID <= 0 {
		return fmt.Errorf("%w: item[%d].chrt_id must be positive", errs.ErrValidation, index)
	}

	if item.TrackNumber == "" {
		return fmt.Errorf("%w: item[%d].track_number is required", errs.ErrValidation, index)
	}
	if utf8.RuneCountInString(item.TrackNumber) > MaxItemTrackNumberLen {
		return fmt.Errorf("%w: item[%d].track_number cannot be longer than %d characters", errs.ErrValidation, index, MaxItemTrackNumberLen)
	}

	if item.Price <= 0 {
		return fmt.Errorf("%w: item[%d].price must be positive", errs.ErrValidation, index)
	}

	if item.Rid == "" {
		return fmt.Errorf("%w: item[%d].rid is required", errs.ErrValidation, index)
	}
	if utf8.RuneCountInString(item.Rid) > MaxItemRidLength {
		return fmt.Errorf("%w: item[%d].rid cannot be longer than %d characters", errs.ErrValidation, index, MaxItemRidLength)
	}
	if !itemRidRegex.MatchString(item.Rid) {
		return fmt.Errorf("%w: item[%d].rid contains invalid characters", errs.ErrValidation, index)
	}

	if item.Name == "" {
		return fmt.Errorf("%w: item[%d].name is required", errs.ErrValidation, index)
	}
	if utf8.RuneCountInString(item.Name) > MaxItemNameLength {
		return fmt.Errorf("%w: item[%d].name cannot be longer than %d characters", errs.ErrValidation, index, MaxItemNameLength)
	}

	if item.Sale < 0 {
		return fmt.Errorf("%w: item[%d].sale cannot be negative", errs.ErrValidation, index)
	}

	if item.Size == "" {
		return fmt.Errorf("%w: item[%d].size is required", errs.ErrValidation, index)
	}
	if utf8.RuneCountInString(item.Size) > MaxItemSizeLength {
		return fmt.Errorf("%w: item[%d].size cannot be longer than %d characters", errs.ErrValidation, index, MaxItemSizeLength)
	}

	if item.TotalPrice <= 0 {
		return fmt.Errorf("%w: item[%d].total_price must be positive", errs.ErrValidation, index)
	}

	if item.NmID <= 0 {
		return fmt.Errorf("%w: item[%d].nm_id must be positive", errs.ErrValidation, index)
	}

	if item.Brand == "" {
		return fmt.Errorf("%w: item[%d].brand is required", errs.ErrValidation, index)
	}
	if utf8.RuneCountInString(item.Brand) > MaxItemBrandLength {
		return fmt.Errorf("%w: item[%d].brand cannot be longer than %d characters", errs.ErrValidation, index, MaxItemBrandLength)
	}

	if item.Status < 0 {
		return fmt.Errorf("%w: item[%d].status cannot be negative", errs.ErrValidation, index)
	}

	return nil
}

func isValidLocale(locale models.LocaleEnum) bool {
	switch locale {
	case models.LocaleEN, models.LocaleRU, models.LocaleES, models.LocaleFR,
		models.LocaleDE, models.LocaleIT, models.LocaleZH, models.LocaleJA,
		models.LocaleKO, models.LocaleAR:
		return true
	default:
		return false
	}
}

func isValidCurrency(currency models.CurrencyEnum) bool {
	switch currency {
	case models.CurrencyUSD, models.CurrencyEUR, models.CurrencyRUB, models.CurrencyGBP,
		models.CurrencyJPY, models.CurrencyCNY, models.CurrencyCAD, models.CurrencyAUD,
		models.CurrencyCHF:
		return true
	default:
		return false
	}
}

func ValidateOrderUID(orderUID string) error {
	if orderUID == "" {
		return fmt.Errorf("order UID cannot be empty")
	}

	if len(orderUID) > 50 {
		return fmt.Errorf("order UID too long")
	}

	if !orderUIDRegex.MatchString(orderUID) {
		return fmt.Errorf("%w: order_uid contains invalid characters", errs.ErrValidation)
	}

	return nil
}
