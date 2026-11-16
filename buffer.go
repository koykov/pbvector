package pbvector

type buffer struct {
	fields  []field
	order   []int
	counts  []int
	starts  []int
	cursor  []int
	gfields []field
	set     map[int]int
}
