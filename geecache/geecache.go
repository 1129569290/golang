package geecache
import (
	"sync"
	"log"
	"fmt"
	"geecache/singlefight"
)

//实现与用户的交互，两种情况
//接收到key，检查是否被缓存
//1 已被缓存，返回缓存值
//2 未被缓存且不必从远程节点获取  调用“回调函数”获取值并添加到缓存，返回缓存值

type Getter interface{
	Get(key string)([]byte, error)
}
type GetterFunc func(key string)([]byte, error)

func (f GetterFunc) Get(key string)([]byte, error){
	return f(key)
}


type Group struct {
	name string
	getter Getter
	mainCache cache
	peers PeerPicker

	loader *singlefight.Group//保证每个key在服务器上只能被访问一次，防止存储该key的服务器压力过大
}
func (g *Group) RegisterPeers(peers PeerPicker){
	if g.peers !=nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers;
}
func (g *Group) load(key string)(value ByteView,err error){
	viewi,err := g.loader.Do(key,func()(interface{},error){
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key);ok{
				if value, err = g.getFromPeer(peer,key);err!=nil{
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer",err)
			}
		}
		return g.getLocally(key)
	})
	if err ==nil {
		return viewi.(ByteView),nil
	}
	return
}

//以上是第五天实现的从peer读取数据






func (g *Group) getFromPeer (peer PeerGetter,key string)(ByteView,error){
	bytes, err :=peer.Get(g.name,key)
	if err !=nil {
		return ByteView{},err
	}
	return ByteView{b : bytes},nil
}




var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group{
	if getter == nil {
		panic("nil Getter")//回调函数为定义
	}
	mu.Lock()
	defer mu.Unlock()
	g :=&Group{
		name :          name,
		getter :       getter ,
		mainCache :    cache{ cacheBytes:cacheBytes },
		loader : &singlefight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g :=groups[name]
	mu.RUnlock()
	return g
} 


//最核心的get方法
func (g *Group) Get(key string) (ByteView,error){
	if key=="" {
		return ByteView{} ,fmt.Errorf("key is required")
	}
	//缓存命中
	if v, ok :=g.mainCache.get(key); ok{
		log.Println("[GeeCache] hit")
		return v, nil
	}
	//缓存未命中

	return g.load(key)
}
//添加了从peer获取的部分


//之前未考虑，直接从本地调用
/*func (g *Group) load(key string)(value ByteView,err error){
	return g.getLocally(key)
}*/
func (g *Group)getLocally(key string) (ByteView,error){
	bytes, err :=g.getter.Get(key)
	if err != nil{
		return ByteView{}, err
	}
	value :=ByteView{b:cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}
func (g *Group) populateCache(key string, value ByteView){
	g.mainCache.add(key, value)
}