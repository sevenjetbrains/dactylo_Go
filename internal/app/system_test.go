package app

import (
	"context"
	"testing"

	"dactylo/internal/core/events"
	"dactylo/internal/store"
)

func TestSystemInfo(t *testing.T) {
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	info, err := NewSystemService(s).Info()
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Version == "" || info.Platform == "" {
		t.Errorf("info incomplète : %+v", info)
	}
	if info.SchemaVersion < 1 {
		t.Errorf("SchemaVersion = %d", info.SchemaVersion)
	}
}

func TestNewRegistersServices(t *testing.T) {
	s, err := store.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	a, err := New(s, events.NewBus())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Shutdown(context.Background())

	if len(a.Services()) == 0 {
		t.Fatal("aucun service enregistré")
	}
	for i, svc := range a.Services() {
		if svc == nil {
			t.Errorf("service %d nil", i)
		}
	}
}
