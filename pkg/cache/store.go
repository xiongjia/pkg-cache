package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Cache struct {
	directory  string
	maxSize    int64
	mu         sync.RWMutex
	accessTime map[string]time.Time
}

func New(directory string, maxSizeMB int) *Cache {
	return &Cache{
		directory:  directory,
		maxSize:    int64(maxSizeMB) * 1024 * 1024,
		accessTime: make(map[string]time.Time),
	}
}

func (c *Cache) FilePath(url string) string {
	hash := sha256.Sum256([]byte(url))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(c.directory, filename)
}

func (c *Cache) Exists(url string) bool {
	path := c.FilePath(url)
	_, err := os.Stat(path)
	return err == nil
}

func (c *Cache) Get(url string) ([]byte, error) {
	path := c.FilePath(url)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.accessTime[path] = time.Now()
	c.mu.Unlock()

	return data, nil
}

func (c *Cache) Put(url string, data []byte) error {
	path := c.FilePath(url)

	// Ensure directory exists
	if err := os.MkdirAll(c.directory, 0755); err != nil {
		return err
	}

	// Check if we need to evict
	if err := c.evictIfNeeded(int64(len(data))); err != nil {
		return err
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	c.mu.Lock()
	c.accessTime[path] = time.Now()
	c.mu.Unlock()

	return nil
}

func (c *Cache) evictIfNeeded(needed int64) error {
	currentSize, err := c.currentSize()
	if err != nil {
		return err
	}

	if currentSize+needed <= c.maxSize {
		return nil
	}

	// Evict oldest files until we have enough space
	c.mu.Lock()
	defer c.mu.Unlock()

	type fileInfo struct {
		path  string
		time  time.Time
		size  int64
	}

	var files []fileInfo
	for path, t := range c.accessTime {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path, t, info.Size()})
	}

	// Sort by access time (oldest first)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].time.Before(files[i].time) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	for _, f := range files {
		if currentSize+needed <= c.maxSize {
			break
		}
		if err := os.Remove(f.path); err == nil {
			delete(c.accessTime, f.path)
			currentSize -= f.size
		}
	}

	return nil
}

func (c *Cache) currentSize() (int64, error) {
	var total int64
	err := filepath.Walk(c.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	if os.IsNotExist(err) {
		return 0, nil
	}
	return total, err
}

func (c *Cache) Stats() (size int64, count int, err error) {
	size, err = c.currentSize()
	if err != nil {
		return 0, 0, err
	}

	c.mu.RLock()
	count = len(c.accessTime)
	c.mu.RUnlock()

	// Re-count actual files
	count = 0
	filepath.Walk(c.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	return size, count, nil
}

func (c *Cache) ShouldCacheByPatterns(url string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := matchPattern(url, pattern)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}
