package cache

import (
	"container/list"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear() // удаляет все ключи
	Add(key, value any)
	AddWithTTL(key, value any, ttl time.Duration) // добавляет ключ со сроком жизни ttl
	Get(key any) (value any, ok bool)
	Remove(key any)
}

type LRUCache struct {
	capacity int                   // Количество ячеек
	cache    map[any]*list.Element // Поле для хранения хэш-таблицы
	queue    *list.List            // Поле двусвязного списка
	mu       sync.Mutex
}

type cacheElem struct {
	key     any       // Поле хранения ключа
	value   any       // Поле хранения кешируемого значения
	expTime time.Time // Время до которого годен элемент кэша
}

// Функция для инициализации LRU кэша, возвращает указатель на экземпляр
func NewLRUCache(cap int) *LRUCache {
	return &LRUCache{
		capacity: cap,                              // Инициализируем количество ячеек в соответсвии с преданным значением
		cache:    make(map[any]*list.Element, cap), // Инициализируем пустую хэш-таблицу
		queue:    list.New(),                       // Инициализируем пустой двухсвязанный список
	}
}

func (c *LRUCache) Cap() int {
	return c.capacity // Так как количество ячеек статично, потокобезопастность не нужна
}

func (c *LRUCache) Len() int {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции
	return c.queue.Len()
}

func (c *LRUCache) Clear() {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции

	c.cache = make(map[any]*list.Element) // Инициализируем пустую хэш-таблицу вместо старой
	c.queue.Init()                        // Инициализируем пустой двухсвязанный список вместо старого
}

func (c *LRUCache) Add(key, value any) {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции

	val, ok := c.cache[key] // Пытаемся найти элемент в хэш-таблице по ключу
	if ok {                 // Случай если переданный ключ уже есть в хэш-таблице
		e := val.Value.(*cacheElem) // Приводим тип к *cacheElem для доступа к полям структуры
		c.queue.MoveToFront(val)    // Перемещаем элемент в начало очереди, так как он недавно использовался
		e.value = value             // Обновляем значение элемента
		e.expTime = time.Time{}     // Сбрасываем время истечения
		return                      // Завершаем метод, так как элемент обновлен
	}
	if c.capacity <= c.queue.Len() { // Случай когда количество элементов очереди равно максимальному количеству ячеек
		c.deleteLeastUsedElem() // Удаляем наименее используемый элемент
	}
	e := &cacheElem{
		key:     key,         // Устанавливаем ключ
		value:   value,       // Устанавливаем значение
		expTime: time.Time{}, // Сбрасываем время истечения
	}
	element := c.queue.PushFront(e) // Добавляем элемент в начало списка
	c.cache[key] = element          // Добавляем элемент в хэш-таблицу для быстрого доступа по ключу
}

func (c *LRUCache) AddWithTTL(key, value any, ttl time.Duration) {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции

	val, ok := c.cache[key] // Пытаемся найти элемент в хэш-таблице по ключу
	if ok {                 // Случай если переданный ключ уже есть в хэш-таблице
		e := val.Value.(*cacheElem) // Приводим тип к *cacheElem для доступа к полям структуры
		c.queue.MoveToFront(val)    // Перемещаем элемент в начало списка, так как он недавно использовался
		e.value = value             // Обновляем значение элемента
		e.expTime = time.Now().Add(ttl)
		return // Завершаем метод, так как элемент обновлен
	}
	if c.capacity == c.queue.Len() { // Случай когда количество элементов очереди равно максимальному количеству ячеек
		c.deleteLeastUsedElem() // Удаляем наименее используемый элемент
	}
	// Создаем новый элемент кэша
	e := &cacheElem{
		key:     key,   // Устанавливаем ключ
		value:   value, // Устанавливаем значение
		expTime: time.Now().Add(ttl),
	}
	element := c.queue.PushFront(e) // Добавляем элемент в начало списка
	c.cache[key] = element          // Добавляем элемент в хэш-таблицу для быстрого доступа по ключу
}

// Метод Get возвращает значение из кэша по ключу и логическое значение указывающее на успешность поиска
func (c *LRUCache) Get(key any) (value any, ok bool) {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции

	val, ok := c.cache[key] // Пытаемся найти элемент в хэш-таблице по ключу
	if !ok {                // Случай когда ключ не найден
		return nil, false
	}
	e := val.Value.(*cacheElem) // Приводим тип к *cacheElem для доступа к полям структуры

	if !e.expTime.IsZero() && time.Now().After(e.expTime) { // Проверяем установено ли время истечения и истек ли срок действия элемента
		c.queue.Remove(val)  // Удаляем элемент из двусвязного списка
		delete(c.cache, key) // Удаляем элемент из хэш-таблицы
		return nil, false
	}
	c.queue.MoveToFront(val) // Перемещаем элемент в начало списка
	return e.value, true     // Возвращаем значение и статус
}

// Метод Remove удаляет значение из кэша по ключу
func (c *LRUCache) Remove(key any) {
	c.mu.Lock()         // Блокируем доступ к кэшу для потокобезопастности
	defer c.mu.Unlock() // Гарантируем разблокировку кэша после выполения метода, даже если произойдет ранний выход из функции

	val, ok := c.cache[key] // Пытаемся найти элемент в хэш-таблице по ключу
	if !ok {                // Случай когда ключ не найден
		return
	}
	c.queue.Remove(val)  // Удаляем значение из кэша по ключу
	delete(c.cache, key) // Удаляет запись из хэш-таблицы
}

// Метод deleteLeastUsedElem удаляет наименее используемый элемент
func (c *LRUCache) deleteLeastUsedElem() {
	back := c.queue.Back() // Берем последний элемент двухсвязанного списка

	if back == nil { // Случай когда список пустой
		return // Завершаем метод, так как список пустой
	}
	c.queue.Remove(back)                         // Удаляем наименее используемый элемент
	delete(c.cache, back.Value.(*cacheElem).key) // Удаляет запись из хэш-таблицы
}
