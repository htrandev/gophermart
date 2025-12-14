package luhn

func Check(value string) bool {
	var sum int
	parity := len(value) % 2
	for i := 0; i < len(value); i++ {
		digit := int(value[i] - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
