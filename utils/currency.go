package utils

// all supported currencies
const (
	NGN = "NGN"
	EUR = "EUR"
	CAD = "CAD"
)

// returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case NGN, EUR, CAD:
		return true
	}
	return false
}