package data

type Stack struct {
	data []int
}

func NewStack(cap int) *Stack {
	return &Stack{data: make([]int, 0, cap)}
}

func (s *Stack) Len() int {
	return len(s.data)
}

func (s *Stack) Push(value int) {
	s.data = append(s.data, value)
}

func (s *Stack) Pop() int {
	r := s.data[s.Len()-1]
	s.data = s.data[:s.Len()-1]
	return r
}

func (s *Stack) Peek() int {
	return s.data[s.Len()-1]
}
