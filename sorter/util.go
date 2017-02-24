package sorter

import "math/rand"

const (
	reqIDLength = 32
)

// newRequestID returns new random request ID
func newRequestID() string {
	bytes := make([]byte, reqIDLength)
	for i := 0; i < reqIDLength; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}
