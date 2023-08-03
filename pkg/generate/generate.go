package generate

import (
	"crypto/rand"
	"math/big"
)

type Generator struct {
	LowerCase bool
	NoDigits bool
	NoSymbols bool
}

const (
	lower = "abcdefghijklmnopqrstuvwxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digit = "0123456789"
	symbol = "~!@#$%^&*()_+`-={}|[]\\:\"<>?,./"
)

func (g *Generator) Generate(size int) (string, error) {
	alphabet := lower

	if !g.LowerCase {
		alphabet += upper
	}

	if !g.NoDigits {
		alphabet += digit
	}

	if !g.NoSymbols {
		alphabet += symbol
	}

	result := ""
	max := big.NewInt(int64(len(alphabet)))

	for i := 0; i < size; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}

		result += string(alphabet[n.Int64()])
	}

	return result, nil
}
