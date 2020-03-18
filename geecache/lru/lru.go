package lru

import "container/list"
//使用的缓存替换算法为LRU,并没有实现同步互斥保护
type Cache struct {
	maxBytes int64               //允许使用的最大内存
	nbytes   int64               //当前已经使用的内存
	ll       *list.List          
	cache    map[string]*list.Element
	OnEvicted func(key string, value Value) //某条记录被移除时的调回函数
}
type  entry struct {
	key string
	value Value//只要实现了Len方法都可以是Value
}
type Value interface {
	Len() int      //返回值所占的内存大小
}
func (c *Cache) Len() int {//实际就是链表的长度
	return c.ll.Len();
}
//cache的构造器
func New(maxBytes int64,onEvicted func(string,Value)) *Cache{
	return &Cache{
		maxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[string]*list.Element),
		//在清除条目时可选并执行
		OnEvicted: onEvicted,
	}
}
// type  Element 本身就含有一个Value接口
//查找功能，返回key对应的value
func (c *Cache) Get(key string) (value Value,ok bool){
	if ele, ok :=c.cache[key]; ok{
		c.ll.MoveToFront(ele)
		kv :=ele.Value.(*entry)//接口断言，返回转化为类型*entry的变量
		return kv.value, true
	} 
	return
}
//删除(将最少使用的替换掉)
func (c *Cache) RemoveOldest(){
	ele :=c.ll.Back()
	if ele!=nil {
		c.ll.Remove(ele)
		kv :=ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -=int64(len(kv.key))+int64(kv.value.Len())
		if c.OnEvicted !=nil {
			c.OnEvicted(kv.key, kv.value)//在队尾删除时被调用
		}
	}
}
//新增
func (c *Cache) Add(key string, value Value) {
	//如果该键已存在，则将之移动到对头
	if ele, ok :=c.cache[key]; ok{
		c.ll.MoveToFront(ele)
		kv :=ele.Value.(*entry)
		c.nbytes+=int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		return
	}
	//如果键不存在
	ele :=c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	c.nbytes+=int64(len(key))+int64(value.Len())
	for c.maxBytes !=0&&c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}