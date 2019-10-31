// 单节点操作
// go run main.go -node_addr="127.0.0.1:9503" -cmd="push" -topic="gmq-topic-1" -push_num=1000
// go run main.go -node_addr="127.0.0.1:9503" -cmd="mpush" -topic="gmq-topic-1" -push_num=1000
// go run main.go -node_addr="127.0.0.1:9503" -cmd="pop" -topic="gmq-topic-1" -pop_num=1000
// go run main.go -node_addr="127.0.0.1:9503" -cmd="pop_loop" -topic="gmq-topic-1"
// go run main.go -node_addr="127.0.0.1:9503" -cmd="ack" -topic="gmq-topic-1" -msg_id="374389276810416130"
// go run main.go -node_addr="127.0.0.1:9503" -cmd="dead" -topic="gmq-topic-1" -pop_num=1000

// 多节点操作
// go run main.go -register_addr="http://127.0.0.1:9595" -cmd="push_by_weight" -topic="gmq-topic-1" -push_num=1000
// go run main.go -register_addr="http://127.0.0.1:9595" -cmd="push_by_avg" -topic="gmq-topic-1" -push_num=1000
// go run main.go -register_addr="http://127.0.0.1:9595" -cmd="push_by_rand" -topic="gmq-topic-1" -push_num=1000
// go run main.go -register_addr="http://127.0.0.1:9595" -cmd="pop_by_weight" -topic="gmq-topic-1"
package main

import (
	"flag"
	"log"
	"time"

	gmq "github.com/wuzhc/gmq-client/client"
)

var (
	cmd          string
	topic        string
	nodeAddr     string
	registerAddr string
	pushNum      int
	popNum       int
	msgId        string
)

func main() {
	flag.StringVar(&cmd, "cmd", "push", "command name")
	flag.StringVar(&topic, "topic", "golang", "topic name")
	flag.StringVar(&nodeAddr, "node_addr", "127.0.0.1:9503", "node address")
	flag.StringVar(&registerAddr, "register_addr", "127.0.0.1:9595", "the address of register")
	flag.StringVar(&msgId, "msg_id", "", "the id of message.")
	flag.IntVar(&pushNum, "push_num", 100, "the number of push, default to 100.")
	flag.IntVar(&popNum, "pop_num", 100, "the number of pop, default to 100.")
	flag.Parse()

	client := gmq.NewClient(nodeAddr, 0)
	start := time.Now()
	switch cmd {
	case "push":
		// 推送消息
		gmq.Example_Produce(client, topic, pushNum)
	case "pop":
		// 拉取消息
		gmq.Example_Consume(client, topic)
	case "pop_loop":
		// 轮询拉取消息
		gmq.Example_Loop_Consume(client, topic)
	case "mpush":
		// 批量推送消息
		gmq.Example_MProduce(client, topic, pushNum)
	case "dead":
		// 拉取死信消息
		gmq.Example_Dead(client, topic, popNum)
	case "ack":
		// 确认已消费
		gmq.Example_Ack(client, topic, msgId)
	case "push_by_weight":
		// 多节点下,按节点权重推送消息
		for i := 0; i < pushNum; i++ {
			c := gmq.GetClientByWeightMode(registerAddr)
			log.Printf("> select node(%v)\n", c.GetAddr())
			gmq.Example_Produce(c, topic, 1)
		}
	case "push_by_rand":
		// 多节点下,按随机模式推送消息
		for i := 0; i < pushNum; i++ {
			c := gmq.GetClientByRandomMode(registerAddr)
			log.Printf("> select node(%v)\n", c.GetAddr())
			gmq.Example_Produce(c, topic, 1)
		}
	case "push_by_avg":
		// 多节点下,按平均模式推送消息
		for i := 0; i < pushNum; i++ {
			c := gmq.GetClientByAvgMode(registerAddr)
			log.Printf("> select node(%v)\n", c.GetAddr())
			gmq.Example_Produce(c, topic, 1)
		}
	case "pop_by_weight":
		// 多节点下,按节点权重消费消息
		for {
			c := gmq.GetClientByWeightMode(registerAddr)
			log.Printf("> select node(%v)\n", c.GetAddr())
			if err := c.Pop(topic); err != nil {
				log.Println(err)
			}

			// receive response
			rtype, data := c.Recv()
			if string(data) == "no message" {
				time.Sleep(1 * time.Second)
			}
			log.Printf("rtype:%v, result:%v\n", rtype, string(data))
		}
	default:
		log.Fatalf("unknown '%v' command.\n", cmd)
	}

	log.Printf("%v in total.\n", time.Now().Sub(start))
}
