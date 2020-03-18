package consistenthash

import (
	"testing"
	"strconv"
)

func TestHashing(t *testing.T){
	hash := New(3,func(key []byte) uint32{
		i, _ :=strconv.Atoi(string(key))
		return uint32(i)
	})
	hash.Add("6","4","2")

	testCasrs := map[string]string{
		"2" : "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	for k, v :=range testCasrs {
		if hash.Get(k) !=v {
			t.Errorf("Asking for %s,should have yield %s", k, v)
		}
	}
	hash.Add("8")

	testCasrs["27"] = "8"
	for k, v := range testCasrs {
		if hash.Get(k) !=v {
			t.Errorf("Asking for %s,should have yielded %s", k ,v)
		} 
	}

}