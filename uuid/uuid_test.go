package uuid_test

import (
	"testing"

	"github.com/socialpoint-labs/bsk/uuid"
)

func TestTimeOrderedUuid(t *testing.T) {
	uu := uuid.New()
	if len(uu) != 36 {
		t.Fatalf("bad UUID size: %s. Must be 36 and is %d", uu, len(uu))
	}
}
