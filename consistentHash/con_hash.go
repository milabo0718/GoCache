package consistenthash

import (
	"errors"
	"hash/crc32"
	"sort"
	"sync"
	"sync/atomic"
)

type Map struct {
	mu sync.RWMutex
	// 哈希环
	keys []int
	// 哈希环到真实节点的映射
	hashMap map[int]string
	// 总的请求数
	totalRequest int64
}

func NewMap() *Map {
	return &Map{
		hashMap:      make(map[int]string),
		totalRequest: 0,
	}
}

func (m *Map) Add(nodes ...string) error {
	if len(nodes) == 0 {
		return errors.New("nodes is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, node := range nodes {
		hash := int(crc32.ChecksumIEEE([]byte(node)))
		m.keys = append(m.keys, hash)
		m.hashMap[hash] = node
	}

	sort.Ints(m.keys)
	return nil
}

func (m *Map) Remove(node string) error {
	if node == "" {
		return errors.New("invalid node")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return nil
}

func (m *Map) Get(key string) string {
	if key == "" {
		return ""
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.keys) == 0 {
		return ""
	}

	hash := int(crc32.ChecksumIEEE([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	if idx == len(m.keys) {
		idx = 0
	}

	atomic.AddInt64(&m.totalRequest, 1)
	return m.hashMap[m.keys[idx]]
}
