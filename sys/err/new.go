package err


import (
	"fmt"
	"time"
)

type myError struct {
	err   string
	time  time.Time
}

func (m *myError) Error() string {
	return fmt.Sprintf("%s %v", m.err, m.time)
}

func New(s string) *myError {
	return &myError{
		err:   s,
		time:  time.Now(),
	}
}

