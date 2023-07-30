package serve

type Pagination struct {
	Current int
	Prev    *int
	Next    *int
	Items   []any
}

func Paginate(current int, max int) Pagination {
	var prev *int
	if current != 1 {
		prevVal := current - 1
		prev = &prevVal
	}

	var next *int
	if current != max {
		nextVal := current + 1
		next = &nextVal
	}

	items := []any{1}

	if current == 1 && max == 1 {
		return Pagination{current, prev, next, items}
	}
	if current > 4 {
		items = append(items, nil)
	}

	r := 2
	r1 := current - r
	r2 := current + r

	for i := maxInt(2, r1); i <= minInt(max, r2); i++ {
		items = append(items, i)
	}

	if r2+1 < max {
		items = append(items, nil)
	}
	if r2 < max {
		items = append(items, max)
	}

	return Pagination{current, prev, next, items}
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
