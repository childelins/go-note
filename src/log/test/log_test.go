package test

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	logger := log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Llongfile)
	logger.Println("hello go")

	logger2 := log.New(os.Stdout, "[xspan-1101] ", log.LstdFlags|log.Llongfile|log.Lmsgprefix)
	logger2.Println("hello go")
}

func TestOutput(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			time.Sleep(500 * time.Microsecond)
			log.Printf("hello %d\n", i)
			wg.Done()
		}(i + 1)
	}

	wg.Wait()
}
