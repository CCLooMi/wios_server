package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 用于测试的哈希表元素数量
const totalElements = 1000000

// 性能对比基准测试
func BenchmarkHashTables(b *testing.B) {
	// 创建新旧哈希表实例
	htNew := NewElasticHashTable(4, 0.75)
	htOld := make(map[string]interface{})
	var mapMutex sync.Mutex

	var wg sync.WaitGroup

	// 新哈希表插入操作测试
	b.Run("New Hash Table Insert", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				htNew.Insert(fmt.Sprintf("key%d", i), i)
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("New Hash Table Insert took: %v", duration)
	})

	// 旧哈希表插入操作测试
	b.Run("Old Hash Table Insert", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				// 锁定旧哈希表
				mapMutex.Lock()
				htOld[fmt.Sprintf("key%d", i)] = i
				mapMutex.Unlock()
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("Old Hash Table Insert took: %v", duration)
	})

	// 新哈希表查找操作测试
	b.Run("New Hash Table Search", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				htNew.Search(fmt.Sprintf("key%d", i))
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("New Hash Table Search took: %v", duration)
	})

	// 旧哈希表查找操作测试
	b.Run("Old Hash Table Search", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				// 锁定旧哈希表
				mapMutex.Lock()
				_, ok := htOld[fmt.Sprintf("key%d", i)]
				mapMutex.Unlock()
				_ = ok // 使用值防止优化
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("Old Hash Table Search took: %v", duration)
	})

	// 新哈希表删除操作测试
	b.Run("New Hash Table Delete", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				htNew.Delete(fmt.Sprintf("key%d", i))
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("New Hash Table Delete took: %v", duration)
	})

	// 旧哈希表删除操作测试
	b.Run("Old Hash Table Delete", func(b *testing.B) {
		b.ResetTimer()
		start := time.Now()
		for i := 0; i < totalElements; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				// 锁定旧哈希表
				mapMutex.Lock()
				delete(htOld, fmt.Sprintf("key%d", i))
				mapMutex.Unlock()
			}(i)
		}
		wg.Wait()
		duration := time.Since(start)
		b.Logf("Old Hash Table Delete took: %v", duration)
	})
}
