package uuid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

// ErrNotEnoughEntropy is returned when the underlying rand libraries are not able to provide enough entropy to create
// a valid UUID
var ErrNotEnoughEntropy = errors.New("Not enough entropy available")

// New returns a time ordered UUID. Top 32 bits are a timestamp, bottom 96 are
// random.
func New() string {
	unix := uint32(time.Now().UTC().Unix())

	b := make([]byte, 12)
	n, err := rand.Read(b)
	if n != len(b) {
		err = ErrNotEnoughEntropy
	}
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		unix, b[0:2], b[2:4], b[4:6], b[6:8], b[8:])
}
