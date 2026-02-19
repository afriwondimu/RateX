package storage

import (
    "sync"
    "time"
)

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
    data map[string]*memoryEntry
    mu   sync.RWMutex
}

type memoryEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        data: make(map[string]*memoryEntry),
    }
}

// Get retrieves a value from memory
func (m *MemoryStorage) Get(key string) (interface{}, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    entry, exists := m.data[key]
    if !exists {
        return nil, false
    }
    
    if time.Now().After(entry.ExpiresAt) {
        // Entry expired, remove it
        m.mu.RUnlock()
        m.mu.Lock()
        delete(m.data, key)
        m.mu.Unlock()
        m.mu.RLock()
        return nil, false
    }
    
    return entry.Value, true
}

// Set stores a value with TTL
func (m *MemoryStorage) Set(key string, value interface{}, ttl time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.data[key] = &memoryEntry{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    }
}

// Delete removes a key
func (m *MemoryStorage) Delete(key string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.data, key)
}

// Clear removes all keys
func (m *MemoryStorage) Clear() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.data = make(map[string]*memoryEntry)
    return nil
}

// Close closes the memory storage
func (m *MemoryStorage) Close() error {
    return m.Clear()
}