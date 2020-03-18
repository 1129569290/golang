package geecache

//为单机搭建httpserver
import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"net/url"
	"io/ioutil"
	"sync"
	"geecache/consistenthash"
)
//第五天第二步：为HTTPPool添加节点选择功能
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)
type HTTPool struct {

	self string       //记录自己的地址
	basePath string   //作为节点间通讯地址的前缀

	mu sync.Mutex //guards peers and httpGetters
	peers  *consistenthash.Map //用来根据具体的key选择节点
	httpGetters map[string]*httpGetter//映射远程节点与对应的httpGetter,
	//每一个远程节点对应一个httpGetter,因为httpGetter与远程节点的地址baseURL有关



}
//实现peerPicker接口
func (p *HTTPool) Set(peers ...string) {//实例化一致性hash算法，并且添加传入的节点
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas,nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _,peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer+p.basePath}
	}
}

//PickPeer picks a peer according to key 
func (p *HTTPool) PickPeer(key string)(PeerGetter,bool){
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer !=p.self {
		p.Log("pick peer %s",peer)
		return p.httpGetters[peer],true
	}
	return nil ,false
}
var _ PeerPicker = (*HTTPool)(nil)




func NewHTTPPool(self string) *HTTPool {
	return &HTTPool{
		self :  self,
		basePath : defaultBasePath,
	}
}

func (p *HTTPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPool) ServeHTTP(W http.ResponseWriter, r *http.Request){
	if !strings.HasPrefix(r.URL.Path, p.basePath){
		panic("HTTPPOOL serving unexpected path:"+r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basename>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) !=2 {
		http.Error(W, "bad requeset",http.StatusBadRequest)
		return
	}

	groupName :=parts[0]

	key :=parts[1]

	group :=GetGroup(groupName)

	if group == nil {
		http.Error(W,"no such group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(W, err.Error(),http.StatusInternalServerError)
		return
	}
	W.Header().Set("Content-type", "application/octet-stream")
	W.Write(view.ByteSlice())
}

//实现HTTP客户端类httpGetter
type httpGetter struct {
	baseURL string
}
func (h *httpGetter) Get(group string, key string)([]byte, error){
	u :=fmt.Sprintf(    //生成格式化的字符串并返回
		"%v%v%v",       //值的默认格式表示
		h.baseURL,
		url.QueryEscape(group),//对字符串进行转码使之可以安全的在URL查询里
		url.QueryEscape(key),
	)
	res, err :=http.Get(u)//向指定的URL发送一个get请求
	if err != nil {
		return nil, err;
	} 
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned : %v",res.Status)
	}

	bytes ,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v",err)
	}
	return bytes,nil
}

var _ PeerGetter = (*httpGetter)(nil)