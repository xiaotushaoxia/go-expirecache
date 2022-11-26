package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	a := NewCache(time.Second+time.Second/3)

	a.Set("1", 1, time.Second)
	a.Set("2", 1, time.Second)
	a.Set("3", 1, time.Second)
	a.Set("4", 1, time.Second)

	time.Sleep(time.Second/2)

	fmt.Println(a.Get("1"))

	time.Sleep(time.Second/3)

	fmt.Println(a.Get("2"))

	time.Sleep(time.Second/2)

	fmt.Println(a.Get("3"))

	time.Sleep(time.Second/2)

	fmt.Println(a.Get("4"))
}


type gs interface {
	Get(key string) (t any, ok bool)
	Set(string, any	, time.Duration)
}

func BenchmarkName(b *testing.B) {
	a := NewCache(time.Second)
	BTest(b, a)
}

type synmp struct {
	mp sync.Map
}

func (m *synmp) Get(k string) (any, bool) {
	return m.mp.Load(k)
}

func (m *synmp) Set(k string, v any, d time.Duration) {
	m.mp.Store(k, v)
}
func BenchmarkName3(b *testing.B) {
	a := synmp{}
	BTest(b, &a)
}

func BTest(b *testing.B, a gs)  {
	fmt.Println(1)
	var wg sync.WaitGroup
	wg.Add(30)

	for j := 0; j < 10; j++ {
		jj := j
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				a.Get(fmt.Sprint(i + jj*100000))
			}
		}()
	}
	for j := 0; j < 10; j++ {
		jj := j
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				a.Set(fmt.Sprint(i+jj*100000), "1", time.Second)
			}
		}()
	}
	for j := 0; j < 10; j++ {
		jj := j
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				a.Set(fmt.Sprint(i+jj*1000000), "1", 1)
			}
		}()
	}
	wg.Wait()
}