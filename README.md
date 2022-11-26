简单的内存cache 只提供Get Set Delete Items  有其他需求要用别的更强大的库

> 删除策略是执行Get Set Delete Items的时候去检查全部的key有没有失效  为了不让检查太过频繁, 有个clearInterval

```go
// 例子
func TestName(t *testing.T) {
    a := NewCache(time.Second+time.Second/3)
    
    a.Set("1", 1, time.Second)
    a.Set("2", 1, time.Second)
    a.Set("3", 1, time.Second)
    a.Set("4", 1, time.Second)
    
    time.Sleep(time.Second/2)
    
    fmt.Println(a.Get("1")) // 1 true
    
    time.Sleep(time.Second/2)
    
    fmt.Println(a.Get("2"))  // 2 true
    
    time.Sleep(time.Second/2)
    
    fmt.Println(a.Get("3"))  //<nil> false
    
    time.Sleep(time.Second/2)
    
    fmt.Println(a.Get("4"))  //<nil> false
}
```