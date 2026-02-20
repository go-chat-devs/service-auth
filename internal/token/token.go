package token

import "crypto/rand"

type Token = [32]byte

func Generate() (t Token) {
	if _, err := rand.Read(t[:]); err != nil {
		panic(err)
	}
	return t
}
