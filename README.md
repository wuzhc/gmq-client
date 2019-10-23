## 单节点使用
```go
package main

import (
	gmq "github.com/wuzhc/gmq-client/client"
)

func main() {
	// 直接连接节点
	client := gmq.NewClient("127.0.0.1:9503", 0)
	// 生产消息
	gmq.Example_Produce(client, "golang")
	// 批量生成消息
	gmq.Example_MProduce(client, "golang")
	// 消费消息
	gmq.Example_Consume(client, "golang")
}
```

## 多节点使用
### 按节点权重投递模式
```go
package main

import (
	"fmt"
	gmq "github.com/wuzhc/gmq-client/client"
)

func main() {
	// 多节点下,按节点权重投递消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByWeightMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}
}
```

### 按平均投递模式
```go
package main

import (
	"fmt"
	gmq "github.com/wuzhc/gmq-client/client"
)

func main() {
	// 多节点下,按平均模式投递消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByAvgMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}
}
```

### 按随机投递模式
```go
package main

import (
	"fmt"
	gmq "github.com/wuzhc/gmq-client/client"
)

func main() {
	// 多节点下,按随机模式投递消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByRandomMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}
}
```