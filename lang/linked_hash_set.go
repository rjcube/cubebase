package lang

import "reflect"

type LinkedHashSet struct {
	m map[interface{}]struct{}
	l []interface{}
}

func NewLinkedHashSet() *LinkedHashSet {
	return &LinkedHashSet{
		m: make(map[interface{}]struct{}),
		l: make([]interface{}, 0),
	}
}

func (s *LinkedHashSet) AddAll(es []interface{}) {
	if nil == es {
		return
	}
	for _, e := range es {
		s.Add(e)
	}
}

func (s *LinkedHashSet) AddStrAll(es []string) {
	if nil == es {
		return
	}
	for _, e := range es {
		s.Add(e)
	}
}

func (s *LinkedHashSet) Add(e interface{}) {
	if _, ok := s.m[e]; !ok {
		s.m[e] = struct{}{}
		s.l = append(s.l, e)
	}
}

func (s *LinkedHashSet) Length() int {
	return len(s.l)
}

// Contains 此方法请谨慎使用，由于该方法使用的是穷举比较，在数据量较大时会导致性能下降
func (s *LinkedHashSet) Contains(e interface{}) bool {
	if nil == e {
		return false
	}
	for _, le := range s.l {
		f := reflect.DeepEqual(e, le)
		if f {
			return true
		}
	}
	return false
}

// Remove 此方法请谨慎使用，由于该方法使用的是穷举比较，在数据量较大时会导致性能下降
func (s *LinkedHashSet) Remove(e interface{}) interface{} {
	if nil == e {
		return nil
	}
	for idx, le := range s.l {
		f := reflect.DeepEqual(e, le)
		if f {
			s.l = RemoveSliceElem(s.l, idx)
			delete(s.m, le)
		}
	}
	return nil
}

func (s *LinkedHashSet) Values() []interface{} {
	return s.l
}

func RemoveSliceElem(slice []interface{}, i int) []interface{} {
	return append(slice[:i], slice[i+1:]...)
}
