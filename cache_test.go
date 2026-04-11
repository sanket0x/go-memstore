package memstore

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
	cache Cache
}

func (suite *CacheTestSuite) SetupTest() {
	// lazy expiry for speed in tests
	suite.cache = NewCache(WithCleanupInterval(0))
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
	sp, ok := suite.cache.(StatsProvider)
	suite.Require().True(ok, "cache must implement StatsProvider")

	suite.cache.Set("a", 1)
	suite.cache.Get("a")       // hit
	suite.cache.Get("a")       // hit
	suite.cache.Get("missing") // miss

	s := sp.Stats()
	suite.Equal(uint64(2), s.Hits)
	suite.Equal(uint64(1), s.Misses)
}

func (suite *CacheTestSuite) TestStatsEvictionsOnMaxKeys() {
	sp, ok := suite.cache.(StatsProvider)
	suite.Require().True(ok)

	c := NewCache(WithCleanupInterval(0), WithMaxKeys(2, PolicyNone))
	defer c.Close()
	sp2 := c.(StatsProvider)

	c.Set("k1", 1)
	c.Set("k2", 2)
	c.Set("k3", 3) // should be rejected

	suite.Equal(2, c.Len())
	suite.Equal(uint64(1), sp2.Stats().Evictions)
	_ = sp
}

func (suite *CacheTestSuite) TestMaxKeysNonePolicyRejectsNewKeys() {
	c := NewCache(WithCleanupInterval(0), WithMaxKeys(2, PolicyNone))
	defer c.Close()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3) // rejected

	suite.Equal(2, c.Len())
	_, ok := c.Get("c")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestMaxKeysAllowsOverwrite() {
	c := NewCache(WithCleanupInterval(0), WithMaxKeys(2, PolicyNone))
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

func (suite *CacheTestSuite) TestBackgroundCleanupRemovesExpiredItems() {
	// build cache with a fast cleanup interval
	c := NewCache(WithCleanupInterval(10 * time.Millisecond))
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
