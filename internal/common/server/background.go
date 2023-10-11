package server

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type BackgroundJobs struct {
	wg sync.WaitGroup
}

func (b *BackgroundJobs) background(fn func()) {
	// Increment the WaitGroup counter.
	b.wg.Add(1)

	// Launch the background goroutine.
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer b.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				logrus.Error(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}
