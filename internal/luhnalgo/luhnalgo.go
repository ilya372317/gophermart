package luhnalgo

func IsValid(number int) bool {
	const baseNumber = 10
	return (number%baseNumber+checksum(number/baseNumber))%baseNumber == 0
}

func checksum(number int) int {
	var luhn int
	const ten, two, zero, nine = 10, 2, 0, 9

	for i := zero; number > zero; i++ {
		cur := number % ten

		if i%two == zero {
			cur *= two
			if cur > nine {
				cur = cur%ten + cur/ten
			}
		}

		luhn += cur
		number /= ten
	}
	return luhn % ten
}
