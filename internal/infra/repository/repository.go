package repository

import (
	"fmt"
	"search_statistics/internal/domain"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Repository struct {
	mu        sync.RWMutex
	GlobalMap map[string]int
	Queue     chan map[string]int
	StopList atomic.Pointer[map[string]bool]

	topCache   atomic.Pointer[[]domain.QueryCount]
	cacheTS    atomic.Int64
}

func New() *Repository {
	q := make(chan map[string]int, 300)
	global := make(map[string]int)
	r := &Repository{GlobalMap: global, Queue: q}
	r.StopList.Store(&map[string]bool{})
	return r
}

func (r *Repository) Update(newData map[string]int) {
	if len(r.Queue) == cap(r.Queue) {
		oldData := <-r.Queue
		r.mu.Lock()
		r.deleteFromGlobal(oldData)
		r.mu.Unlock()
	}
	r.Queue <- newData
	r.mu.Lock()
	r.addToGlobal(newData)
	r.mu.Unlock()

	r.cacheTS.Store(0) 
}

func (r *Repository) addToGlobal(newData map[string]int) {
	for key, value := range newData {
		r.GlobalMap[key] += value
	}
}

func (r *Repository) deleteFromGlobal(newData map[string]int) {
	for key, value := range newData {
		r.GlobalMap[key] -= value
		if r.GlobalMap[key] <= 0 {
            delete(r.GlobalMap, key)
        }
	}
}

func (r *Repository) AddInStopList(word string) error {
	word = strings.ToLower(strings.TrimSpace(word))
	old := r.StopList.Load()
	newMap := make(map[string]bool, len(*old)+1)
	for k, v := range *old {
		newMap[k] = v
	}
	if (newMap[word]) {
		return fmt.Errorf("this word already in stop list")
	}
	newMap[word] = true
	r.StopList.Store(&newMap)
	r.cacheTS.Store(0)
	return nil
}

func (r *Repository) DeleteFromStopList(word string) error {
	word = strings.ToLower(strings.TrimSpace(word))
	old := r.StopList.Load()
	newMap := make(map[string]bool, len(*old))
	for k, v := range *old {
		newMap[k] = v
	}
	if !newMap[word] {
		return fmt.Errorf("this word is not in the stop list")
	}
	delete(newMap, word)
	r.StopList.Store(&newMap)
	r.cacheTS.Store(0)
	return nil
}

func (r *Repository) GetTopCached(n int) []domain.QueryCount {
	if time.Since(time.UnixMilli(r.cacheTS.Load())).Milliseconds() < 500 {
		if cached := r.topCache.Load(); cached != nil {
			return *cached
		}
	}
	top := r.GetTop(n)
	r.topCache.Store(&top)
	r.cacheTS.Store(time.Now().UnixMilli())
	return top
}


func (r *Repository) GetTop(n int) ([]domain.QueryCount) {
	r.mu.RLock()
	pairs := make([]domain.QueryCount, 0, len(r.GlobalMap))
	for q, c := range r.GlobalMap {
		if r.hasWordInStopList(q) {
			continue
		}
		pairs = append(pairs, domain.QueryCount{Query: q, Count: c})
	}
	r.mu.RUnlock()
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	if n >= len(pairs) {
        return pairs
    }
    return pairs[:n]

}

func (r *Repository) hasWordInStopList(query string) bool {
	slPtr := r.StopList.Load()
	if slPtr == nil {
		return false
	}
	sl := *slPtr
    tokens := strings.FieldsFunc(query, func(char rune) bool {
        return char == ' ' || char == ',' || char == '.' || char == '!' || char == '?' || char == '-' || char == '*' || char == '_'
    })
    for _, token := range tokens {
        token = strings.TrimSpace(token)
        if token == "" {
            continue
        }
        if sl[strings.ToLower(token)] {
            return true
        }
    }
    return false
}