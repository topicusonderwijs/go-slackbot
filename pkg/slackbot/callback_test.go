package slackbot

import (
	"sync"
	"testing"
	"time"
)

func TestGCSweepRemovesExpired(t *testing.T) {
	CallbackStorage = sync.Map{}

	fresh := NewCallback()
	stale := NewCallback()
	stale.Created = time.Now().Add(-2 * callbackTTL)

	gcSweep()

	if _, err := FindCallback(fresh.Id.String()); err != nil {
		t.Errorf("fresh callback should be kept, got: %v", err)
	}
	if _, err := FindCallback(stale.Id.String()); err == nil {
		t.Errorf("stale callback should have been removed")
	}
}

func TestCallbackConcurrentAccess(t *testing.T) {
	CallbackStorage = sync.Map{}

	const goroutines = 50

	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Go(func() {
			cb := NewCallback()
			cb.Set("n", i)
			cb.Set("name", "value")

			if _, err := FindCallback(cb.Id.String()); err != nil {
				t.Errorf("callback should be findable: %v", err)
			}
			_ = cb.GetInt("n")
			_ = cb.GetString("name")
			gcSweep()
		})
	}
	wg.Wait()
}
