package lru

import(
	"reflect"
	"testing"
)
type String string
func (d String) Len() int{//实现了value接口
		return len(d)
}
//测试新增函数
func TestGet(t *testing.T){
	lru :=New(int64(0),nil)//后面一个参数是OnEvicted
	lru.Add("key1",String("1234"))
	//返回的v是一个Value(接口)类型的变量
	if v, ok :=lru.Get("key1"); !ok||string(v.(String)) !="1234"{
		t.Fatal("未将key1 he 1234 匹配")
	}
	if _,ok :=lru.Get("key2"); ok{
		t.Fatal("匹配到key2就离谱了")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 :="value1", "value2", "v3"
	cap :=len(k1+k2+v1+v2)
	lru :=New(int64(cap),nil)
	lru.Add(k1,String(v1))
	lru.Add(k2,String(v2))
	lru.Add(k3,String(v3))

	if _,ok :=lru.Get("key1"); ok ||lru.Len() !=2 {
		t.Fatal("移除失败")
	}
}
func TestOnEvicted(t *testing.T) {
	keys :=make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys,key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))
	expect :=[]string{"key1","k2"}
	if !reflect.DeepEqual( expect, keys){
		t.Fatal("调回函数失败")
	}
	
}
func TestAdd(t *testing.T) {
	lru :=New(int64(0),nil)
	lru.Add("key",String("1"))
	lru.Add("key",String("111"))
	if lru.nbytes !=int64(len("key")+len("111")){
		t.Fatal("期望6 但是得到",lru.nbytes)
	}

}