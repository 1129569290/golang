package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)
//计算hash
type Hash func(data []byte)uint32

type Map struct {
	hash Hash
	replicas int  //虚拟节点倍数
	keys []int    //sorted  hash环
	hashMap map[int]string //虚拟节点与真是节点的映射表，键是虚拟节点的哈希值，值是真是节点的名称
}

func New(replicas int, fn Hash) *Map{//允许自定义节点倍数和Hash函数
	m := &Map{
		replicas : replicas ,
		hash     :  fn      ,
		hashMap  : make(map[int]string),
	 }
	 if m.hash ==nil {
		 m.hash = crc32.ChecksumIEEE
	 }
	 return m;
}

//添加真实节点/机器的Add()方法
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i<m.replicas; i++ {
			hash :=int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)//对hash环就行排序
}

//实现选择节点的Get()方法

//从hash距离最近的节点取值
func (m *Map) Get(key string) string {
	if len(m.keys) == 0{
		return ""
	}
	hash :=int(m.hash([]byte(key)))

	idx :=sort.Search(len(m.keys),func(i int) bool{
		return m.keys[i]>=hash
	})//找到匹配的下标
	return m.hashMap[m.keys[idx % len(m.keys)]]//映射得到真实的节点
}