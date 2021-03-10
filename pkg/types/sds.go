package types

import (
	"github.com/jin1ming/Gedis/pkg/utils"
	"strings"
)

type SdsInterface interface {
	Append(s string)
	//Free()
	ToString() string
}

func NewSds(str string) *Sds {
	return &Sds{
		free: 0,
		buf:  utils.StringToBytes(str),
	}
}

type Sds struct {
	free int
	buf  []byte
}

func (s *Sds) Append(str string) {
	if len(str) > s.free {
		var buf []byte
		if len(s.buf)+len(str) > 16384 {
			buf = make([]byte, len(s.buf), len(s.buf)+len(str)+16384)
		} else {
			buf = make([]byte, len(s.buf), (len(s.buf)+len(str))*2)
		}
		copy(buf, s.buf)
		s.buf = buf
	}
	s.buf = append(s.buf, utils.StringToBytes(str)...)
	s.free = cap(s.buf) - len(s.buf)
}

func (s *Sds) String() string {
	return utils.BytesToString(&s.buf)
}

/* Compare two sds strings s1 and s2 with strings.Compare().
 *
 * Return value:
 *
 *     positive if s1 > s2.
 *     negative if s1 < s2.
 *     0 if s1 and s2 are exactly the same binary string.
 *
 * If two strings share exactly the same prefix, but one of the two has
 * additional characters, the longer string is considered to be greater than
 * the smaller one. */
func SdsCompare(s1, s2 Sds) int {
	return strings.Compare(s1.String(), s2.String())
}
