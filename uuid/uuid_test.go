package uuid

import (
	"testing"
)

func TestTimeOrderedUuid(t *testing.T) {
	uuid := New()
	if len(uuid) != 36 {
		t.Fatalf("bad UUID size: %s. Must be 36 and is %d", uuid, len(uuid))
	}
}
