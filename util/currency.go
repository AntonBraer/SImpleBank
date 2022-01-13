package util


const (
	USD = "USD"
	EUR ="EUR"
)

func IsCurrencySupport(currency string) bool{
	switch currency{
	case USD, EUR:
		return true
	}
	return false
}