package repository_test

import (
	"fmt"
	"search_statistics/internal/domain"
	"search_statistics/internal/infra/repository"
	"sync"
	"testing"
	"time"
)

func assertTopEqual(t *testing.T, got, want []domain.QueryCount) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Query != want[i].Query || got[i].Count != want[i].Count {
			t.Errorf("index %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestNew(t *testing.T) {
	r := repository.New()
	if r == nil {
		t.Fatal("expected non-nil repository")
	}
	if cap(r.Queue) != 300 {
		t.Errorf("expected queue capacity 300, got %d", cap(r.Queue))
	}
	if r.StopList.Load() == nil {
		t.Error("expected initialized stop list")
	}
}

func TestUpdateAndAggregation(t *testing.T) {
	r := repository.New()

	r.Update(map[string]int{"iphone 15": 10})
	r.Update(map[string]int{"iphone 15": 5, "samsung s24": 2})

	top := r.GetTop(10)
	assertTopEqual(t, top, []domain.QueryCount{
		{Query: "iphone 15", Count: 15},
		{Query: "samsung s24", Count: 2},
	})
}

func TestSlidingWindowEviction(t *testing.T) {
	r := repository.New()

	for i := 0; i < 300; i++ {
		r.Update(map[string]int{fmt.Sprintf("sec_%d", i): 1})
	}

	if len(r.GetTop(400)) != 300 {
		t.Fatal("expected 300 items after filling the window")
	}

	r.Update(map[string]int{"new_query": 1})

	top := r.GetTop(400)
	if len(top) != 300 {
		t.Fatalf("expected 300 items after eviction, got %d", len(top))
	}

	foundOld, foundNew := false, false
	for _, q := range top {
		if q.Query == "sec_0" {
			foundOld = true
		}
		if q.Query == "new_query" {
			foundNew = true
		}
	}
	if foundOld {
		t.Error("oldest entry 'sec_0' should be evicted from window")
	}
	if !foundNew {
		t.Error("new entry 'new_query' should be present in window")
	}
}

func TestStopListCRUD(t *testing.T) {
	r := repository.New()

	err := r.AddInStopList("  APPLE  ")
	if err != nil {
		t.Fatalf("failed to add stop word: %v", err)
	}

	err = r.AddInStopList("apple")
	if err == nil {
		t.Error("expected error when adding duplicate stop word")
	}

	err = r.DeleteFromStopList("Apple")
	if err != nil {
		t.Fatalf("failed to remove stop word: %v", err)
	}

	err = r.DeleteFromStopList("apple")
	if err == nil {
		t.Error("expected error when removing non-existent stop word")
	}
}

func TestStopListFilteringInTop(t *testing.T) {
	r := repository.New()
	r.Update(map[string]int{
		"iphone pro max":    20,
		"samsung galaxy":    10,
		"pixel phone":       5,
		"pro accessories":   8, 
	})

	r.AddInStopList("pro")

	top := r.GetTop(10)

	for _, q := range top {
		if q.Query == "iphone pro max" || q.Query == "pro accessories" {
			t.Errorf("query containing stop word '%s' should be filtered", q.Query)
		}
	}

	assertTopEqual(t, top, []domain.QueryCount{
		{Query: "samsung galaxy", Count: 10},
		{Query: "pixel phone", Count: 5},
	})
}

func TestGetTopN_LargerThanMap(t *testing.T) {
	r := repository.New()
	r.Update(map[string]int{"a": 1, "b": 2})

	top := r.GetTop(10)
	if len(top) != 2 {
		t.Fatalf("expected 2 items, got %d", len(top))
	}
}

func TestTopCacheTTL(t *testing.T) {
	r := repository.New()
	r.Update(map[string]int{"x": 100, "y": 50})

	res1 := r.GetTopCached(2)

	res2 := r.GetTopCached(2)
	if &res1[0] != &res2[0] {
		t.Error("expected same cached slice pointer for TTL < 500ms")
	}

	time.Sleep(600 * time.Millisecond)

	res3 := r.GetTopCached(2)
	if &res1[0] == &res3[0] {
		t.Error("expected new slice pointer after TTL expiration")
	}
	assertTopEqual(t, res3, res1)
}

func TestCacheInvalidationOnUpdate(t *testing.T) {
	r := repository.New()
	r.Update(map[string]int{"a": 1})

	res1 := r.GetTopCached(1)
	pointer1 := &res1[0]

	r.Update(map[string]int{"b": 2})

	res2 := r.GetTopCached(2)
	pointer2 := &res2[0]

	if pointer1 == pointer2 {
		t.Error("cache should be invalidated after Update(), pointer should change")
	}
	if len(res2) != 2 {
		t.Fatalf("expected 2 items after update, got %d", len(res2))
	}
}

func TestConcurrencySafety(t *testing.T) {
	r := repository.New()
	var wg sync.WaitGroup
	iterations := 5000

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			r.Update(map[string]int{fmt.Sprintf("concurrent_%d", i%100): 1})
		}
	}()

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = r.GetTopCached(10)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			word := fmt.Sprintf("stop_%d", id)
			_ = r.AddInStopList(word)
			time.Sleep(10 * time.Millisecond)
			_ = r.DeleteFromStopList(word)
		}(i)
	}

	wg.Wait()
}