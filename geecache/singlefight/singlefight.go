//防止缓存击穿
//缓存击穿：一个存在的key,在缓存过期的一刻，同时有大量的请求，
//这些请求会击穿db

package singlefight

import "sync"

type call struct { //代表正在进行中，或者已经结束的请求，使用sync.WaitGroup避免重入
	we sync.WaitGroup
	val interface{}
	err error
}
type Group struct {  //管理不同key的请求
	mu sync.Mutex
	m map[string]*call
}
//Do的作用，就死针对相同的key,无论Do被调用多少次，函数fn都只会被调用一次，
//等待fn调用结束了，返回值或者错误
func (g *Group) Do(key string, fn func()(interface{}, error))(interface{},error){
	g.mu.Lock()
	if g.m==nil {
		g.m=make(map[string]*call)
	}
	if c, ok := g.m[key];ok {
		g.mu.Unlock()
		c.we.Wait()        //如果请求正在进行中，则等待
		return c.val, c.err//请求结束，返回结果
	}
	c :=new(call)
	c.we.Add(1)            //发起请求前加锁
	g.m[key]=c			   //添加到g.m 表明key已经有对应的请求在处理
	g.mu.Unlock()


	c.val, c.err = fn()   //调用fn 发起请求
	c.we.Done()           //请求结束
	g.mu.Lock()
	delete(g.m, key)      //更新g.m
	g.mu.Unlock()

	return c.val, c.err

}