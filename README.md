# ip_location

### 介绍
根据ip获取location（国家、城市、区号等）信息
需要结合QQwry纯真IP数据库

### 获取纯真IP库
访问 http://www.cz88.net 下载纯真IP库

### 使用
```
    go get github.com/lbw114007/ip_location
```

### demo
```
package main

import (
	"github.com/lbw114007/ip_location"
	"log"
)

func main() {
	IPDict := ip_location.NewIPDict()
    //载入IP字典
	err := IPDict.Load("./qqwry.dat")
	if err != nil {
		log.Fatalln(err)
	}
    //查询IP
	res, err := IPDict.FindIP("127.0.0.1")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(res)
}
```