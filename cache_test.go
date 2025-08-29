package memstore

import (
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
	suite.cache.Set("foo", "bar", 0)

	val, ok := suite.cache.Get("foo")
	suite.True(ok)
	suite.Equal("bar", val)
}

func (suite *CacheTestSuite) TestGetNonExistentKey() {
	_, ok := suite.cache.Get("doesnotexist")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestExpiry() {
	suite.cache.Set("temp", "gone", 20*time.Millisecond)
	time.Sleep(40 * time.Millisecond)
	_, ok := suite.cache.Get("temp")
	suite.False(ok)
}

func (suite *CacheTestSuite) TestDeleteAndExists() {
	suite.cache.Set("k1", "v1", 0)
	suite.True(suite.cache.Exists("k1"))

	suite.cache.Delete("k1")
	suite.False(suite.cache.Exists("k1"))
}

func (suite *CacheTestSuite) TestKeysAndLen() {
	suite.cache.Set("user:1", "a", 0)
	suite.cache.Set("user:2", "b", 0)
	suite.cache.Set("order:1", "x", 0)

	keys := suite.cache.Keys("user:*")
	suite.Len(keys, 2)

	suite.Equal(3, suite.cache.Len())
}

func (suite *CacheTestSuite) TestBackgroundCleanupRemovesExpiredItems() {
	// build cache with a fast cleanup interval
	c := NewCache(WithCleanupInterval(10 * time.Millisecond))
	defer c.Close()

	c.Set("t", "v", 20*time.Millisecond)
	// item should exist initially
	v, ok := c.Get("t")
	suite.True(ok)
	suite.Equal("v", v)

	// wait past expiry + cleanup interval
	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("t")
	suite.False(ok)
}
