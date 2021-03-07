package types

import (
	"github.com/jin1ming/Gedis/pkg/utils"
)

type SDSInterface interface {
	Append(s string)
	//Free()
	ToString() string
}

func NewSDS(str string) *SDS {
	return &SDS{
		free: 0,
		buf:  []byte(str),
	}
}

type SDS struct {
	free int
	buf []byte
}

func (s *SDS)Append(str string) {
	if len(str) > s.free {
		var buf []byte
		if len(s.buf) + len(str) > 16384 {
			buf = make([]byte, 0, len(s.buf) + len(str) + 16384)
		} else {
			buf = make([]byte, 0, (len(s.buf) + len(str)) * 2)
		}
		copy(buf, s.buf)
		s.free = cap(buf) - len(buf)
		s.buf = buf
	} else {
		s.buf = append(s.buf, utils.StringToBytes(str)...)
	}
}

func (s *SDS)String() string {
	return utils.BytesToString(&s.buf)
}
