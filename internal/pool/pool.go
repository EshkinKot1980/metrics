// Модуль pool реализует временное хранилище для параметризованных типов данных.
package pool

import "sync"

// Resettable определяет интерфейс для объектов способных сбрасывать свое состояние.
type Resettable interface {
	Reset()
}

// Pool предоставляет потоко-безопасный пул объектов типа Т.
// T должен соответсвовать интерфейсу Resettable.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// Создает новый пул объектов типа T.
// f функция создающая объект типа Т, вызывается когда пул пустой.
func New[T Resettable](f func() T) *Pool[T] {

	p := Pool[T]{}
	p.pool.New = func() any {
		return f()
	}
	return &p
}

// Отдает объект из пула.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Возвращает объект в пул очищая его состояние.
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
