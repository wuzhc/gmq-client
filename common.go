package gmq

import (
	"sync"
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}

func RemoveSliceElem(s *[]interface{}, index int) interface{} {
	if len(*s) == 0 {
		return nil
	}

	e := (*s)[index]
	if index == 0 {
		*s = nil
	} else {
		*s = append((*s)[0:index], (*s)[index:]...)
	}

	return e
}
