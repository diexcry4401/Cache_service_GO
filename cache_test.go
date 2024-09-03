package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestLRUCache_AddAndGet(t *testing.T) {
	cache := NewLRUCache(2) // Создаем кэш с емкостью 2 элемента

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")
	cache.Add("key2", "value4")

	// Проверка получения существующих ключей
	if val, ok := cache.Get("key1"); !ok || val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	if val, ok := cache.Get("key2"); !ok || val != "value4" {
		t.Errorf("Expected value2, got %v", val)
	}

	// Добавление нового ключа, должен вытеснить key1
	cache.Add("key3", "value3")

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be evicted")
	}

	if val, ok := cache.Get("key2"); !ok || val != "value4" {
		t.Errorf("Expected value2, got %v", val)
	}

	if val, ok := cache.Get("key3"); !ok || val != "value3" {
		t.Errorf("Expected value3, got %v", val)
	}
}

func TestLRUCache_AddWithTTL(t *testing.T) {
	cache := NewLRUCache(2) // Создаем кэш с емкостью 2 элемента

	cache.AddWithTTL("key1", "value1", 1*time.Second)
	cache.AddWithTTL("key2", "value2", 3*time.Second)
	cache.AddWithTTL("key2", "value4", 5*time.Second)

	// Добавление нового ключа, должен вытеснить key1
	cache.AddWithTTL("key5", "value4", 1*time.Second)

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be evicted")
	}

	// Ожидаем, что оба ключа доступны сразу после добавления
	if val, ok := cache.Get("key5"); !ok || val != "value4" {
		t.Errorf("Expected value1, got %v", val)
	}

	if val, ok := cache.Get("key2"); !ok || val != "value4" {
		t.Errorf("Expected value2, got %v", val)
	}

	// Ждем, пока key1 истечет
	time.Sleep(2 * time.Second)

	// Проверка получения существующих ключей и несуществующих ключей
	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be expired")
	}

	if _, ok := cache.Get("key5"); ok {
		t.Errorf("Expected key1 doesn't exist")
	}

	if val, ok := cache.Get("key2"); !ok || val != "value4" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestLRUCache_Remove(t *testing.T) {
	cache := NewLRUCache(2) // Создаем кэш с емкостью 2 элемента

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	cache.Remove("key1")

	// Проверка получения существующих ключей и несуществующих ключей
	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed")
	}

	if _, ok := cache.Get("key3"); ok {
		t.Errorf("Expected key3 doesn't exist")
	}

	if val, ok := cache.Get("key2"); !ok || val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestLRUCache_Remove_From_Empty(t *testing.T) {
	cache := NewLRUCache(2) // Создаем кэш с емкостью 2 элемента

	cache.Remove("key1")

	if val, _ := cache.Get("key1"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	if val, _ := cache.Get("key2"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	if val, _ := cache.Get("key3"); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache(2) // Создаем кэш с емкостью 2 элемента

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache to be empty after clear")
	}

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed after clear")
	}

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("Expected key2 to be removed after clear")
	}
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCache(10) // Создаем кэш с емкостью 10 элементов

	var wg sync.WaitGroup
	numGoroutines := 20 // Задаем кол-во желаемых горутин

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			cache.Add(key, i) // Добавляем значение в кэш

			if val, ok := cache.Get(key); !ok || val != i {
				t.Errorf("Goroutine %d: expected %d, got %v", i, i, val)
			}

			if i%2 == 0 { // Каждую вторую итерацию удаляем элемент
				cache.Remove(key)
				val, ok := cache.Get(key)
				if ok {
					t.Errorf("Goroutine %d: expected nil, got %v", i, val)
				}
			}
		}(i)
	}

	wg.Wait() // Ожидаем завершения всех горутин

	// В конце проверяем, что количество элементов в кэше не превышает его емкость
	if cache.Len() > cache.Cap() {
		t.Errorf("Cache length exceeded capacity: got %d, want <= %d", cache.Len(), cache.Cap())
	}
}
