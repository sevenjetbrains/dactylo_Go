package events

import (
	"sync"
	"testing"
)

func TestPublishReachesMatchingSubscribers(t *testing.T) {
	b := NewBus()
	var got []string
	b.Subscribe(SessionCompleted, func(e Event) { got = append(got, "session:"+e.Name) })
	b.Subscribe(WordMistyped, func(e Event) { got = append(got, "word:"+e.Name) })
	b.SubscribeAll(func(e Event) { got = append(got, "all:"+e.Name) })

	b.Publish(Event{Name: SessionCompleted})

	want := []string{"session:SessionCompleted", "all:SessionCompleted"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	b := NewBus()
	calls := 0
	unsub := b.Subscribe(TextImported, func(Event) { calls++ })
	b.Publish(Event{Name: TextImported})
	unsub()
	unsub() // idempotent
	b.Publish(Event{Name: TextImported})
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestPayloadIsDelivered(t *testing.T) {
	b := NewBus()
	type payload struct{ Word string }
	var got payload
	b.Subscribe(WordMistyped, func(e Event) { got = e.Payload.(payload) })
	b.Publish(Event{Name: WordMistyped, Payload: payload{Word: "être"}})
	if got.Word != "être" {
		t.Errorf("payload = %+v", got)
	}
}

func TestHandlerMayPublishAndUnsubscribe(t *testing.T) {
	b := NewBus()
	var unsub func()
	nested := 0
	unsub = b.Subscribe(SessionCompleted, func(Event) {
		unsub()
		b.Publish(Event{Name: WordMistyped})
	})
	b.Subscribe(WordMistyped, func(Event) { nested++ })

	b.Publish(Event{Name: SessionCompleted}) // ne doit pas bloquer
	b.Publish(Event{Name: SessionCompleted})
	if nested != 1 {
		t.Errorf("nested = %d, want 1", nested)
	}
}

func TestConcurrentUse(t *testing.T) {
	b := NewBus()
	var mu sync.Mutex
	count := 0
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsub := b.Subscribe(TextImported, func(Event) {
				mu.Lock()
				count++
				mu.Unlock()
			})
			b.Publish(Event{Name: TextImported})
			unsub()
		}()
	}
	wg.Wait()
	if count == 0 {
		t.Error("aucun événement reçu")
	}
}
