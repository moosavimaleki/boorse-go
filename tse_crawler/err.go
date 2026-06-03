package tse_crawler

import "fmt"

type notFoundHistory struct {
	id  uint64
	min int
}

func (e *notFoundHistory) Error() string {
	return fmt.Sprintf("%d[%d] Not Found History", e.id, e.min)
}
