package memstore

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	cache Cache[any]
}

func (suite *CacheTestSuite) SetupTest() {
	suite.cache = NewCache[any](WithCleanupInterval(0))
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) TearDownTest() {
	if suite.cache != nil {
		suite.cache.Close()
	}
}

func (suite *CacheTestSuite) TestSetAndGet() {
	suite.cache.Set("foo", "bar")

	val, ok := suite.cache.Get("foo")
	suite.True(ok)
	suite.Equal("bar", val)
}

func (suite *CacheTestSuite) TestGetNonExistentKey() {
	_, ok := suite.cache.Get("doesnotexist")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestExpiry() {
	suite.cache.SetWithDuration("temp", "gone", 20*time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	_, ok := suite.cache.Get("temp")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestDeleteAndExists() {
	suite.cache.Set("k1", "v1")
	suite.True(suite.cache.Exists("k1"))

	suite.cache.Delete("k1")
	suite.False(suite.cache.Exists("k1"))
}

func (suite *CacheTestSuite) TestKeysAndLen() {
	suite.cache.Set("user:1", "a")
	suite.cache.Set("user:2", "b")
	suite.cache.Set("order:1", "x")

	keys := suite.cache.Keys("user:*")
	suite.Len(keys, 2)

	suite.Equal(3, suite.cache.Len())
}

func (suite *CacheTestSuite) TestDeleteReturnsTrue() {
	suite.cache.Set("x", 1)
	suite.True(suite.cache.Delete("x"))
}

func (suite *CacheTestSuite) TestDeleteReturnsFalse() {
	suite.False(suite.cache.Delete("nonexistent"))
}

func (suite *CacheTestSuite) TestStatsHitsAndMisses() {
	statsOpt, stats := WithStats()
	c := NewCache[any](WithCleanupInterval(0), statsOpt)
	defer c.Close()

	c.Set("a", 1)
	c.Get("a")       // hit
	c.Get("a")       // hit
	c.Get("missing") // miss

	s := stats.Snapshot()
	suite.Equal(uint64(2), s.Hits)
	suite.Equal(uint64(1), s.Misses)
}

func (suite *CacheTestSuite) TestStatsEvictionsOnMaxKeys() {
	statsOpt, stats := WithStats()
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(2, PolicyNone), statsOpt)
	defer c.Close()

	c.Set("k1", 1)
	c.Set("k2", 2)
	suite.ErrorIs(c.Set("k3", 3), ErrCacheFull)

	suite.Equal(2, c.Len())
	suite.Equal(uint64(1), stats.Snapshot().Evictions)
}

func (suite *CacheTestSuite) TestMaxKeysNonePolicyRejectsNewKeys() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(2, PolicyNone))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // rejected

	suite.Equal(2, c.Len())
	_, ok := c.Get("c")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestMaxKeysAllowsOverwrite() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(2, PolicyNone))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("a", 99) // overwrite — must succeed

	suite.Equal(2, c.Len())
	val, ok := c.Get("a")
	suite.True(ok)
	suite.Equal(99, val)
}

func (suite *CacheTestSuite) TestConcurrentAccess() {
	const goroutines = 100
	const ops = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key"
			for j := 0; j < ops; j++ {
				suite.cache.Set(key, id)
				suite.cache.Get(key)
				suite.cache.Exists(key)
				suite.cache.Delete(key)
			}
		}(i)
	}
	wg.Wait()
}

func (suite *CacheTestSuite) TestLRUEvictsLeastRecentlyUsed() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(3, PolicyLRU))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Access "a" and "b" — making "c" the least recently used
	c.Get("a")
	c.Get("b")

	// Adding "d" should evict "c"
	c.Set("d", 4)

	suite.Equal(3, c.Len())
	_, ok := c.Get("c")
	suite.False(ok, "c should have been evicted")
	_, ok = c.Get("d")
	suite.True(ok)
}

func (suite *CacheTestSuite) TestLRUOverwriteUpdatesRecency() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(2, PolicyLRU))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("a", 99) // overwrite — makes "a" most recent, "b" least recent

	// Adding "c" should evict "b"
	c.Set("c", 3)

	suite.Equal(2, c.Len())
	_, ok := c.Get("b")
	suite.False(ok, "b should have been evicted")
	val, ok := c.Get("a")
	suite.True(ok)
	suite.Equal(99, val)
}

func (suite *CacheTestSuite) TestLFUEvictsLeastFrequentlyUsed() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(3, PolicyLFU))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Access "a" 3 times, "b" 2 times, "c" 1 time (just the insert)
	c.Get("a")
	c.Get("a")
	c.Get("a")
	c.Get("b")
	c.Get("b")

	// Adding "d" should evict "c" (lowest frequency = 1)
	c.Set("d", 4)

	suite.Equal(3, c.Len())
	_, ok := c.Get("c")
	suite.False(ok, "c should have been evicted (lowest freq)")
	_, ok = c.Get("d")
	suite.True(ok)
}

func (suite *CacheTestSuite) TestLFUTieBreaksByRecency() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(2, PolicyLFU))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2) // both at freq=1; "a" is older (inserted first = least recent)

	// Adding "c" should evict "a" (same freq, least recently inserted)
	c.Set("c", 3)

	suite.Equal(2, c.Len())
	_, ok := c.Get("a")
	suite.False(ok, "a should have been evicted (same freq, least recent)")
	_, ok = c.Get("b")
	suite.True(ok)
}

func (suite *CacheTestSuite) TestLRUConcurrentAccess() {
	c := NewCache[any](WithCleanupInterval(0), WithMaxKeys(50, PolicyLRU))
	defer c.Close()

	// Pre-fill
	for i := 0; i < 50; i++ {
		c.Set(fmt.Sprintf("k%d", i), i)
	}

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				c.Set(fmt.Sprintf("new-%d-%d", id, j), id)
				c.Get(fmt.Sprintf("k%d", j%50))
			}
		}(i)
	}
	wg.Wait()
}

func (suite *CacheTestSuite) TestBackgroundCleanupRemovesExpiredItems() {
	// build cache with a fast cleanup interval
	c := NewCache[any](WithCleanupInterval(10 * time.Millisecond))
	defer c.Close()

	c.SetWithDuration("t", "v", 20*time.Millisecond)
	// item should exist initially
	v, ok := c.Get("t")
	suite.True(ok)
	suite.Equal("v", v)

	// wait past expiry + cleanup interval
	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("t")
	suite.False(ok)
}
