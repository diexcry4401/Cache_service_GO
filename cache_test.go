package cache

import (
	"testing"
	"time"
)

func TestLRUCache_AddAndGet(t *testing.T) {
	cache := NewLRUCache(2)

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
	cache := NewLRUCache(2)

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

	if _, ok := cache.Get("key5"); ok {
		t.Errorf("Expected key1 to be expired")
	}

	if val, ok := cache.Get("key2"); !ok || val != "value4" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestLRUCache_Remove(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Add("key1", "value1")
	cache.Add("key2", "value2")

	cache.Remove("key1")

	if _, ok := cache.Get("key1"); ok {
		t.Errorf("Expected key1 to be removed")
	}

	if _, ok := cache.Get("key3"); ok {
		t.Errorf("Expected key3 doesn't")
	}

	if val, ok := cache.Get("key2"); !ok || val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestLRUCache_Remove_From_Empty(t *testing.T) {
	cache := NewLRUCache(2)

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
	cache := NewLRUCache(2)

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
