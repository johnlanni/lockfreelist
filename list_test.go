package lockfreelist

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestLockfreeList_Stack(t *testing.T) {
	var list LockfreeList
	for i := 0; i < 100; i++ {
		list.PushFront(i)
	}
	for i := 99; i >= 0; i-- {
		e := list.PopFront()
		if e == nil {
			t.Fatal("got nil")
		}
		if e.Value.(int) != i {
			t.Fatalf("expect %d bug got %v", i, e.Value)
		}
	}
}

func TestLockfreeList_Queue(t *testing.T) {
	var list LockfreeList
	for i := 0; i < 100; i++ {
		list.PushBack(i)
	}
	for i := 0; i < 100; i++ {
		e := list.PopFront()
		if e == nil {
			t.Fatal("got nil")
		}
		if e.Value.(int) != i {
			t.Fatalf("expect %d bug got %v", i, e.Value)
		}
	}
}

func TestLockfreeList_Concurrency(t *testing.T) {
	var list LockfreeList
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(2)
		go func(index int) {
			list.PushBack(index)
			wg.Done()
		}(i)
		go func(index int) {
			list.PushFront(index)
			wg.Done()
		}(i)
	}
	wg.Wait()
	p := list.begin
	for p != nil {
		fmt.Printf("addr:%x p:%+v\n", p, (*Element)(p))
		p = (*Element)(p).next
	}
	cm := map[int]int{}
	var mutex sync.Mutex
	var err error
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e := list.PopFront()
			fmt.Printf("pop e:%+v\n", e)
			if e == nil {
				err = errors.New("got nil")
				return
			}
			mutex.Lock()
			defer mutex.Unlock()
			cm[e.Value.(int)] = cm[e.Value.(int)] + 1
		}()
	}
	wg.Wait()
	if err != nil {
		t.Fatalf("got err: %v", err)
	}
	for key, value := range cm {
		if value != 2 {
			t.Fatalf("key:%d expect value: %d, got: %d", key, 2, value)
		}
	}
}
