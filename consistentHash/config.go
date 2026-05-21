package consistenthash

import "hash/crc32"

type Config struct {
	// 虚拟节点的数量
	VirtualNodeCount int

	// 最小虚拟节点数量
	MinVirtualNodeCount int

	// 最大虚拟节点数量
	MaxVirtualNodeCount int

	// 哈希函数
	HashFunc func(data []byte) uint32

	// 负载均衡阈值
	LoadBalanceThreshold float64
}

var DefaultConfig = &Config{
	VirtualNodeCount:     100,
	MinVirtualNodeCount:  10,
	MaxVirtualNodeCount:  1000,
	HashFunc:             crc32.ChecksumIEEE,
	LoadBalanceThreshold: 0.4,
}
