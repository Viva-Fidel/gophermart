package utils

// Проверяет, является ли номер заказа корректным
func ValidLuhn(number string) bool {
	if number == "" {
		return false
	}
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		c := number[i]
		if c < '0' || c > '9' {
			return false
		}
		digit := int(c - '0')
		if alt {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alt = !alt
	}
	return sum%10 == 0
}
