package consistenthash

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Map struct {
	mu sync.RWMutex
	// 配置
	config *Config
	// 哈希环
	keys []int
	// 哈希环到真实节点的映射
	hashMap map[int]string
	// 节点到虚拟节点数量的映射
	nodeVirtualCount map[string]int
	// 节点负载
	nodeLoad map[string]float64
	// 总的请求数
	totalRequest int64
}

func NewMap() *Map {
	m := &Map{
		config:           DefaultConfig,
		hashMap:          make(map[int]string),
		nodeVirtualCount: make(map[string]int),
		nodeLoad:         make(map[string]float64),
	}
	m.startLoadBalancer()
	return m
}

func (m *Map) Add(nodes ...string) error {
	if len(nodes) == 0 {
		return errors.New("nodes is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, node := range nodes {
		if node == "" {
			continue
		}

		m.addNode(node, m.config.VirtualNodeCount)
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

	cnt := m.nodeVirtualCount[node]
	if cnt == 0 {
		return errors.New("node not found")
	}

	m.removeNode(node)
	sort.Ints(m.keys)
	return nil
}

func (m *Map) Get(key string) string {
	if key == "" {
		return ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.config.HashFunc([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	if idx == len(m.keys) {
		idx = 0
	}

	node := m.hashMap[m.keys[idx]]
	count := m.nodeLoad[node]
	m.nodeLoad[node] = count + 1
	atomic.AddInt64(&m.totalRequest, 1)

	return node
}

func (m *Map) addNode(node string, virtualNodeCount int) {
	for i := 0; i < virtualNodeCount; i++ {
		hash := int(m.config.HashFunc([]byte(fmt.Sprintf("%s-%d", node, i))))
		m.keys = append(m.keys, hash)
		m.hashMap[hash] = node
	}
	m.nodeVirtualCount[node] = virtualNodeCount
}

func (m *Map) removeNode(node string) {
	cnt := m.nodeVirtualCount[node]
	for i := 0; i < cnt; i++ {
		hash := int(m.config.HashFunc([]byte(fmt.Sprintf("%s-%d", node, i))))
		delete(m.hashMap, hash)
		for j := 0; j < len(m.keys); j++ {
			if m.keys[j] == hash {
				m.keys = append(m.keys[:j], m.keys[j+1:]...)
				break
			}
		}
	}

	delete(m.nodeVirtualCount, node)
	delete(m.nodeLoad, node)
}

func (m *Map) checkAndRebalance() {
	// 样本不足，不进行负载均衡
	if atomic.LoadInt64(&m.totalRequest) < 1000 {
		return
	}

	m.mu.RLock()
	if len(m.nodeVirtualCount) == 0 {
		m.mu.RUnlock()
		return
	}

	// 计算负载情况
	avgLoad := float64(m.totalRequest) / float64(len(m.nodeVirtualCount))
	var maxDiff float64

	for _, count := range m.nodeLoad {
		diff := math.Abs(float64(count) - avgLoad)
		if diff/avgLoad > maxDiff {
			maxDiff = diff / avgLoad
		}
	}
	m.mu.RUnlock()

	// 如果负载不均衡度超过阈值，调整虚拟节点
	if maxDiff > m.config.LoadBalanceThreshold {
		m.rebalanceNodes()
	}
}

func (m *Map) rebalanceNodes() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.nodeVirtualCount) == 0 {
		return
	}

	avgLoad := float64(m.totalRequest) / float64(len(m.nodeVirtualCount))
	// 重新计算每个节点的虚拟节点数量
	for node := range m.nodeVirtualCount {
		curNodeCnt := m.nodeVirtualCount[node]
		loadRatio := m.nodeLoad[node] / avgLoad

		var newCnt int
		if loadRatio > 1 {
			newCnt = int(float64(curNodeCnt) / loadRatio)
		} else {
			newCnt = int(float64(curNodeCnt) * (2 - loadRatio))
		}

		if newCnt < m.config.MinVirtualNodeCount {
			newCnt = m.config.MinVirtualNodeCount
		}

		if newCnt > m.config.MaxVirtualNodeCount {
			newCnt = m.config.MaxVirtualNodeCount
		}

		if newCnt != curNodeCnt {
			m.removeNode(node)
			m.addNode(node, newCnt)
		}
	}

	for node := range m.nodeLoad {
		m.nodeLoad[node] = 0
	}
	atomic.StoreInt64(&m.totalRequest, 0)

	// 重新排序
	sort.Ints(m.keys)
}

func (m *Map) GetLoad(node string) map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	load := make(map[string]float64)
	total := atomic.LoadInt64(&m.totalRequest)
	if total == 0 {
		return load
	}

	for n, count := range m.nodeLoad {
		load[n] = count / float64(total)
	}

	return load
}

func (m *Map) startLoadBalancer() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			m.checkAndRebalance()
		}
	}()
}
