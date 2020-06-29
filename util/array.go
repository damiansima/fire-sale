package util

func Sum(array []float32) float32 {
	var result float32
	for _, v := range array {
		result += v
	}
	return result
}

func Equals(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
