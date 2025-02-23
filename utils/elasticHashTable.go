package utils

import (
	"fmt"
	"sync"
)

type HashTableElement struct {
	key   string
	value interface{}
}

type ElasticHashTable struct {
	mu                  sync.RWMutex // 保护哈希表操作的读写锁
	table               []*HashTableElement
	size                int
	filled              int
	loadFactorThreshold float64
}

func NewElasticHashTable(initialSize int, loadFactorThreshold float64) *ElasticHashTable {
	// init the table
	table := make([]*HashTableElement, initialSize)
	return &ElasticHashTable{
		table:               table,
		size:                initialSize,
		filled:              0,
		loadFactorThreshold: loadFactorThreshold,
	}
}

// calculate the hash value
func (ht *ElasticHashTable) hash(key string) int {
	hashValue := 0
	for i := 0; i < len(key); i++ {
		hashValue = (hashValue*31 + int(key[i])) % ht.size
	}
	return hashValue
}

func (ht *ElasticHashTable) checkResize() {
	loadFactor := float64(ht.filled) / float64(ht.size)
	if loadFactor >= ht.loadFactorThreshold {
		ht.resize()
	}
}

func (ht *ElasticHashTable) resize() {
	newSize := ht.size * 2
	newTable := make([]*HashTableElement, newSize)

	for i := 0; i < ht.size; i++ {
		if ht.table[i] != nil {
			element := ht.table[i]
			index := ht.hashForResize(element.key, newSize)
			probes := 0
			for probes < newSize {
				newIndex := (index + probes) % newSize
				if newTable[newIndex] == nil {
					newTable[newIndex] = element
					break
				}
				probes++
			}
		}
	}

	ht.table = newTable
	ht.size = newSize
}

func (ht *ElasticHashTable) hashForResize(key string, newSize int) int {
	hashValue := 0
	for i := 0; i < len(key); i++ {
		hashValue = (hashValue*31 + int(key[i])) % newSize
	}
	return hashValue
}

func (ht *ElasticHashTable) Insert(key string, value interface{}) {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	ht.checkResize()

	index := ht.hash(key)
	probes := 0

	// 使用非贪婪策略插入
	// 通过探测多个位置插入
	for probes < ht.size {
		i := (index + probes) % ht.size
		if ht.table[i] == nil {
			ht.table[i] = &HashTableElement{key, value}
			ht.filled++
			return
		}
		probes++
	}
}

func (ht *ElasticHashTable) Search(key string) (interface{}, bool) {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	index := ht.hash(key)
	probes := 0

	// 通过探测多个位置查找元素
	for probes < ht.size {
		i := (index + probes) % ht.size
		if ht.table[i] == nil {
			return nil, false
		}
		if ht.table[i].key == key {
			return ht.table[i].value, true
		}
		probes++
	}

	return nil, false
}

func (ht *ElasticHashTable) Delete(key string) bool {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	index := ht.hash(key)
	probes := 0

	// 通过探测多个位置删除元素
	for probes < ht.size {
		i := (index + probes) % ht.size
		if ht.table[i] == nil {
			return false
		}
		if ht.table[i].key == key {
			ht.table[i] = nil
			ht.filled--
			return true
		}
		probes++
	}

	return false
}

func (ht *ElasticHashTable) Print() {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	for i, element := range ht.table {
		if element != nil {
			fmt.Printf("Index %d: {Key: %s, Value: %v}\n", i, element.key, element.value)
		}
	}
}
