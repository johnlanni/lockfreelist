package lockfreelist

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

type Element struct {
	next  unsafe.Pointer
	Value interface{}
}

type LockfreeList struct {
	begin unsafe.Pointer
	end   unsafe.Pointer
}

func (l *LockfreeList) PushFront(v interface{}) {
	e := &Element{next: atomic.LoadPointer(&l.begin), Value: v}
	if atomic.CompareAndSwapPointer(&l.end, nil, unsafe.Pointer(e)) {
		atomic.StorePointer(&l.begin, unsafe.Pointer(e))
		return
	}
	for !atomic.CompareAndSwapPointer(&l.begin, e.next, unsafe.Pointer(e)) {
		runtime.Gosched()
		atomic.StorePointer(&e.next, l.begin)
	}
}

func (l *LockfreeList) PushBack(v interface{}) {
	e := &Element{next: nil, Value: v}
	var end unsafe.Pointer
	for {
		end = atomic.LoadPointer(&l.end)
		if end == nil {
			if atomic.CompareAndSwapPointer(&l.end, nil, unsafe.Pointer(e)) {
				atomic.StorePointer(&l.begin, unsafe.Pointer(e))
				return
			}
		} else {
			next := atomic.LoadPointer(&(*Element)(end).next)
			if end == atomic.LoadPointer(&l.end) {
				if next == nil {
					if atomic.CompareAndSwapPointer(&(*Element)(end).next, next, unsafe.Pointer(e)) {
						break
					}
				} else {
					atomic.CompareAndSwapPointer(&l.end, end, next)
				}
			}
		}
		runtime.Gosched()
	}
	atomic.CompareAndSwapPointer(&l.end, end, unsafe.Pointer(e))
}

func (l *LockfreeList) PopFront() *Element {
	var begin unsafe.Pointer
	for {
		begin = atomic.LoadPointer(&l.begin)
		if begin == nil {
			return nil
		}
		if !atomic.CompareAndSwapPointer(&l.begin, begin, unsafe.Pointer((*Element)(begin).next)) {
			runtime.Gosched()
			continue
		} else if (*Element)(begin).next == nil {
			atomic.StorePointer(&l.end, nil)
		}
		break
	}
	return (*Element)(begin)
}
