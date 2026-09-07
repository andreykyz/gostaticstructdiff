package nested

import (
	"fmt"
	"time"
)

type GGID string

// gostaticstructdiff
type (
	MetaString map[string]string

	MetaTime map[int]time.Time
	ID       int
)

func (i *MetaTime) MarshalText() []byte {
	return []byte("ggg")
}

func (i *ID) String() string {
	return fmt.Sprintf("ID(%d)", i)
}
