package geecache

// 缓存值的抽象与封装
type ByteView struct{
	b []byte//存储真实的缓存值,entry里面的value值，例如字符串和图片 byte是一个ascii码
}
//返回view的的长度
func (v ByteView) Len() int {
	return len(v.b)
}

//返回一个拷贝
func (v ByteView) ByteSlice() []byte{
	return cloneBytes(v.b)
}

func cloneBytes(b []byte) []byte {
	c:=make([]byte, len(b))
	copy(c,b)
	return c
}
func (v ByteView) String() string {
	return string(v.b)
}