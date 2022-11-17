// Package crypto contains shared cryptographic functions
package crypto

import "crypto/sha1"

func HashIt(val string) []byte {

	hashFunc := sha1.New()

	hashedVal := hashFunc.Sum([]byte(val))

	return hashedVal
}
