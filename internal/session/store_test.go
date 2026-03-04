package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	t.Parallel()

	store := NewStore()
	created, err := store.Create(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if created.ID != "session-1" {
		t.Fatalf("unexpected session id: %s", created.ID)
	}

	got, ok, err := store.Get(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected session to exist")
	}
	if got.ID != created.ID {
		t.Fatalf("unexpected get result: %s", got.ID)
	}
}

func TestStoreCreateDuplicate(t *testing.T) {
	t.Parallel()

	store := NewStore()
	if _, err := store.Create(context.Background(), "duplicate"); err != nil {
		t.Fatalf("first create returned error: %v", err)
	}

	if _, err := store.Create(context.Background(), "duplicate"); !errors.Is(err, ErrSessionExists) {
		t.Fatalf("expected ErrSessionExists, got: %v", err)
	}
}

func TestStoreConcurrentCreate(t *testing.T) {
	t.Parallel()

	store := NewStore()
	const workers = 32

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := fmt.Sprintf("session-%d", idx)
			if _, err := store.Create(context.Background(), id); err != nil {
				t.Errorf("create %s returned error: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < workers; i++ {
		id := fmt.Sprintf("session-%d", i)
		if _, ok, err := store.Get(context.Background(), id); err != nil || !ok {
			t.Fatalf("expected %s to exist, ok=%v err=%v", id, ok, err)
		}
	}
}

func TestStoreRespectsCanceledContext(t *testing.T) {
	t.Parallel()

	store := NewStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := store.Create(ctx, "session"); err == nil {
		t.Fatal("expected create to fail with canceled context")
	}

	if _, _, err := store.Get(ctx, "session"); err == nil {
		t.Fatal("expected get to fail with canceled context")
	}
}
