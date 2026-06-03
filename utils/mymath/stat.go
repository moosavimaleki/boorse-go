package mymath

func MeanF32(s []float32) float32 {
	var total float32 = 0
	for _, value := range s {
		total += value
	}
	return total / float32(len(s))
}
func MeanF64(s []float64) float64 {
	var total float64 = 0
	for _, value := range s {
		total += value
	}
	return total / float64(len(s))
}
