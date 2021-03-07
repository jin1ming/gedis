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
		buf:  utils.StringToBytes(str),
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
			buf = make([]byte, len(s.buf), len(s.buf) + len(str) + 16384)
		} else {
			buf = make([]byte, len(s.buf), (len(s.buf) + len(str)) * 2)
		}
		copy(buf, s.buf)
		s.buf = buf
	}
	s.buf = append(s.buf, utils.StringToBytes(str)...)
	s.free = cap(s.buf) - len(s.buf)
}

func (s *SDS)String() string {
	return utils.BytesToString(&s.buf)
}
