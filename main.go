package main

import (
	"fmt"

	gmq "github.com/wuzhc/gmq-client/client"
)

func main() {
	client := gmq.NewClient("127.0.0.1:9503", 0)
	gmq.Example_Produce(client, "golang")
	gmq.Example_MProduce(client, "golang")
	gmq.Example_Consume(client, "golang")

	// 多节点下,按节点权重推送消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByWeightMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}

	// 多节点下,按随机模式推送消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByRandomMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}

	// 多节点下,按平均模式推送消息
	for i := 0; i < 10; i++ {
		c := gmq.GetClientByAvgMode()
		fmt.Println(c.GetAddr())
		gmq.Example_Produce(c, "golang")
	}
}
