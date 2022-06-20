package rt

import "math/rand"

// RandFloatRange Returns a random real in [min,max).
func RandFloatRange(min, max float64) float64 {
	return min + (max-min)*RandFloat()
}

func RandFloat() float64 {
	return rand.Float64()
}
