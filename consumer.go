// 消费者
// 功能:
// 	- 添加回调函数,处理任务
package gmq

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Consumer struct {
	handers   map[string]*ConsumerConnHandler
	limitChan chan int
	exitChan  chan struct{}
	topic     string
	cfg       *ConsumerCfg
	mux       sync.RWMutex
	wg        WaitGroupWrapper
	event     *Event
	running   int32
}

func NewConsumer(topic string, cfg *ConsumerCfg) *Consumer {
	return &Consumer{
		topic:     topic,
		exitChan:  make(chan struct{}),
		limitChan: make(chan int, cfg.MaxConcurrent),
		handers:   make(map[string]*ConsumerConnHandler),
		event:     &Event{},
		cfg:       cfg,
	}
}

func (c *Consumer) Run() {
	if !atomic.CompareAndSwapInt32(&c.running, 0, 1) {
		log.Fatalln("Start consumer failed")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigs
		atomic.AddInt32(&c.running, -1)
		close(c.exitChan)
		c.closeAllConnHandler()
	}()

	c.wg.Wrap(c.timer)
	<-c.exitChan
}

func (c *Consumer) Stop() {
	c.wg.Wait()
	log.Println("Consumer had stop.")
}

// 关闭所有连接处理器
func (c *Consumer) closeAllConnHandler() {
	for _, h := range c.handers {
		h.Do(h.exit)
	}
}

// 定时器,定时轮询注册中心节点,更新节点(新增,删除)
func (c *Consumer) timer() {
	defer log.Println("consumer timer had exit.")

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-c.exitChan:
			return
		case <-ticker.C:
			nodes, err := GetNodes(c.cfg.RegisterAddr)
			if err != nil {
				log.Println(err)
				continue
			}

			for addr, h := range c.handers {
				isOffLine := true // 标识是否下线
				for k, node := range nodes {
					if node.TcpAddr == addr {
						isOffLine = false
						nodes = RemoveNode(nodes, k)
						break
					}
				}
				// 节点下线处理,关闭连接处理器
				if isOffLine {
					h.Do(h.exit)
				}
			}

			// 新节点
			for _, node := range nodes {
				if err := c.ConnectToNode(node); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (c *Consumer) OnMessage(f RespMsgFunc) {
	c.event.OnMessage = f
}

func (c *Consumer) OnJob(f RespMsgFunc) {
	c.event.OnJob = f
}

func (c *Consumer) OnError(f RespMsgFunc) {
	c.event.OnError = f
}

// 连接到注册中心
func (c *Consumer) ConnectToRegister() error {
	if len(c.cfg.RegisterAddr) == 0 {
		return errors.New("Register addr is empty")
	}

	nodes, err := GetNodes(c.cfg.RegisterAddr)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := c.ConnectToNode(node); err != nil {
			return err
		}
	}

	return nil
}

// 连接到节点
func (c *Consumer) ConnectToNode(node *Node) error {
	if c.event.OnJob == nil || c.event.OnError == nil {
		return errors.New("Must be set callback function before connect to node.")
	}

	conn, err := net.Dial("tcp", node.TcpAddr)
	if err != nil {
		return err
	}

	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.handers[node.TcpAddr]; ok {
		return errors.New(fmt.Sprintf("addr %s has exist\n", node.TcpAddr))
	}

	handler := &ConsumerConnHandler{
		Addr:        node.TcpAddr,
		consumer:    c,
		conn:        conn,
		nextJobChan: make(chan int, 1),
		exitChan:    make(chan struct{}),
	}

	c.handers[node.TcpAddr] = handler
	c.wg.Wrap(handler.Run)

	return nil
}

// 连接到指定节点
func (c *Consumer) ConnectSpecifiedNode(tcpAddr string, weight int) error {
	node := &Node{
		TcpAddr: tcpAddr,
		Weight:  weight,
	}

	return c.ConnectToNode(node)
}
