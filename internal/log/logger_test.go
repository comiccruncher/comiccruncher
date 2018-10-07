package log

import (
	"fmt"
	"testing"
)

// Test for race conditions.
func TestLogger(t *testing.T) {
	ch := make(chan int, 100)
	for i := 0; i < 100; i++ {
		go func(i int) {
			Logger(Cerebro)
			Logger(Web)
			ch <- i
		}(i)
	}
	for i := 0; i < 100; i++ {
		fmt.Println(<-ch)
	}
}

func BenchmarkLogger(b *testing.B) {
	ch := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			Logger(Cerebro)
			Logger(Web)
			ch <- i
		}()
	}
	for i := 0; i < b.N; i++ {
		fmt.Println(<-ch)
	}
}
