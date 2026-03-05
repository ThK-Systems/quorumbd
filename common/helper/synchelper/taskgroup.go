// Package synchelper provides helper functions for working with the sync package
package synchelper

import "sync"

type TaskGroup struct {
	wg sync.WaitGroup
}

func (g *TaskGroup) Go(fn func()) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		// TODO: Maybe use later to recover panics
		//        defer func() {
		//            if r := recover(); r != nil {
		//                log.Printf("panic: %v", r)
		//            }
		//        }()

		fn()
	}()
}

func (g *TaskGroup) Wait() {
	g.wg.Wait()
}
