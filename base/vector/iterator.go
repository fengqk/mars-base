package vector

type Iterator[T any] struct {
	vec   *Vector[T]
	index int32
}

func (v *Vector[T]) Iterator() Iterator[T] {
	return Iterator[T]{vec: v, index: -1}
}

func (v *Iterator[T]) Next() bool {
	if v.index < v.vec.count {
		v.index++
	}
	return v.vec.WithInRange(v.index)
}

func (v *Iterator[T]) Prev() bool {
	if v.index >= 0 {
		v.index--
	}
	return v.vec.WithInRange(v.index)
}

func (v *Iterator[T]) Value() T {
	return v.vec.Get(v.index)
}

func (v *Iterator[T]) Index() int32 {
	return v.index
}

func (v *Iterator[T]) Begin() {
	v.index = -1
}

func (v *Iterator[T]) End() {
	v.index = v.vec.count
}

func (v *Iterator[T]) First() bool {
	v.Begin()
	return v.Next()
}

func (v *Iterator[T]) Last() bool {
	v.End()
	return v.Prev()
}
