package queue

import "container/list"

type Queue interface {
	Front() *list.Element
	Len() int
	Add(interface{})
	Remove()
}

type queueImpl struct {
	*list.List // anonymous fields or embedding
}

func (q *queueImpl) Add(v interface{}) {
	q.PushBack(v)
}

func (q *queueImpl) AddV2(v interface{}) int {
	q.Add(v)
	return q.Len() - 1
}

func (q *queueImpl) Remove() {
	re := q.Front()
	q.List.Remove(re)
}

func New() Queue {
	return &queueImpl{list.New()}
}
