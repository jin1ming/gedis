package data_struct

type Set struct {
	s map[string]bool
}

func NewSet() *Set {
	return &Set{
		make(map[string]bool),
	}
}

func (s *Set) Add(value string) {
	s.s[value] = true
}

func (s *Set) Has(value string) bool {
	if _, ok := s.s[value]; ok {
		return true
	}
	return false
}

func (s *Set) GetAllBytes() [][]byte {
	values := make([][]byte, 0, len(s.s))
	for v := range s.s {
		values = append(values, []byte(v))
	}
	return values
}

func (s *Set) Delete(value string) {
	if _, ok := s.s[value]; !ok {
		return
	}
	delete(s.s, value)
}

func (s *Set) Length() int {
	return len(s.s)
}
