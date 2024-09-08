package buid

import (
	"math/rand"
	"time"
)

const (
	base32Alphabet    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	buidLength        = 24
	randomCharsLength = 17
)

func GenerateBUID() string {
	// Pre-allocate byte slice for result
	result := make([]byte, buidLength)

	// Convert the current Unix timestamp to base32
	now := time.Now().UTC().Unix()
	pos := buidLength - randomCharsLength

	for now > 0 {
		pos--
		result[pos] = base32Alphabet[now&31]
		now /= 32
	}

	// Generate the random part directly into the result slice
	for i := buidLength - randomCharsLength; i < buidLength; i++ {
		// n, _ := rand.Int(rand.Reader, big.NewInt(32)) // Ignoring the error as before
		// result[i] = base32Alphabet[n.Int64()]
		result[i] = base32Alphabet[rand.Intn(32)]
	}

	return string(result[pos:])
}
