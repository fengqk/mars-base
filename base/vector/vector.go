package vector

import (
	"log"
)

const (
	VectorBlockSize = 16
)

type (
	Vector[T any] struct {
		count int32
		size  int32
		array []T
	}
)

func assert(x bool, y string) {
	if bool(x) == false {
		log.Printf("\nFatal :{%s}", y)
	}
}

func (v *Vector[T]) shift(pos int32) {
	assert(pos <= v.count, "vector shift out of bounds")

	if pos == v.size {
		v.resize(v.count + 1)
	} else {
		v.count++
	}

	for i := v.count - 1; i > pos; i-- {
		v.array[i] = v.array[i-1]
	}
}

func (v *Vector[T]) resize(count int32) {
	blocks := count / VectorBlockSize
	if count%VectorBlockSize != 0 {
		blocks++
	}
	v.count = count
	v.size = blocks * VectorBlockSize
	array := make([]T, v.size+1)
	copy(array, v.array)
	v.array = array
}

func (v *Vector[T]) incr() {
	if v.count == v.size {
		v.resize(v.count + 1)
	} else {
		v.count++
	}
}

func (v *Vector[T]) decr() {
	assert(v.count > 0, "vector decr count is zero")
	v.count--
}

func (v *Vector[T]) PushFront(val T) {
	v.shift(0)
	v.array[0] = val
}

func (v *Vector[T]) PushBack(val T) {
	v.incr()
	v.array[v.count-1] = val
}

func (v *Vector[T]) PopFront() {
	assert(v.count > 0, "Vector popFront count is zero")
	v.Erase(0)
}

func (v *Vector[T]) PopBack() {
	assert(v.count > 0, "Vector popBack count is zero")
	v.decr()
}

func (v *Vector[T]) WithInRange(index int32) bool {
	return index >= 0 && index < v.count
}

func (v *Vector[T]) Erase(pos int32) {
	assert(pos < v.count, "Vector erase out of bounds")
	if pos < v.count-1 {
		copy(v.array[pos:v.count], v.array[pos+1:v.count])
	}
	v.count--
}

func (v *Vector[T]) Front() T {
	assert(v.count > 0, "Vector front count is zero")
	return v.array[0]
}

func (v *Vector[T]) Back() T {
	assert(v.count > 0, "Vector back count is zero")
	return v.array[v.count-1]
}

func (v *Vector[T]) Empty() bool {
	return v.count == 0
}

func (v *Vector[T]) Size() int32 {
	return v.size
}

func (v *Vector[T]) Count() int32 {
	return v.count
}

func (v *Vector[T]) Clear() {
	v.count = 0
}

func (v *Vector[T]) Get(pos int32) T {
	assert(pos < v.count, "Vector get out of bounds")
	return v.array[pos]
}

func (v *Vector[T]) Values() []T {
	return v.array[0:v.count]
}

func (v *Vector[T]) Swap(i, j int32) {
	v.array[i], v.array[j] = v.array[j], v.array[i]
}
