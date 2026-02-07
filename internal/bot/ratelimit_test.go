package bot

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	lim := newRateLimiter(2, 200*time.Millisecond)
	key := int64(123)

	if ok, _ := lim.Allow(key); !ok {
		t.Fatal("expected first allow")
	}
	if ok, _ := lim.Allow(key); !ok {
		t.Fatal("expected second allow")
	}
	if ok, retry := lim.Allow(key); ok {
		t.Fatal("expected deny")
	} else if retry <= 0 {
		t.Fatalf("expected positive retry, got %v", retry)
	}

	time.Sleep(220 * time.Millisecond)
	if ok, _ := lim.Allow(key); !ok {
		t.Fatal("expected allow after window")
	}
}
