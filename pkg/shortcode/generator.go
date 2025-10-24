package shortcode

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	base62Chars   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	defaultLength = 6
)

type Generator struct {
	length int
}

func NewGenerator(length int) *Generator {
	if length <= 0 {
		length = defaultLength
	}
	return &Generator{length: length}
}

func (g *Generator) Generate() string {
	var sb strings.Builder
	sb.Grow(g.length)

	for i := 0; i < g.length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(62))
		if err != nil {
			idx = big.NewInt(int64(i * 17 % 62))
		}
		sb.WriteByte(base62Chars[idx.Int64()])
	}

	return sb.String()
}

func (g *Generator) Validate(code string) bool {
	if len(code) != g.length {
		return false
	}

	for _, char := range code {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}

	return true
}
