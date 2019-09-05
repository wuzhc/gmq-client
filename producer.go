package gmq

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type Producter struct {
	weight   int32
	mux      sync.RWMutex
	wg       WaitGroupWrapper
	handlers map[string]*ProducterConnHander
	exitChan chan struct{}
	event    *Event
	running  int32
	cfg      *ProducerCfg
}

func NewProducter(cfg *ProducerCfg) *Producter {
	p := &Producter{
		exitChan: make(chan struct{}),
		handlers: make(map[string]*ProducterConnHander),
		event:    &Event{},
		cfg:      cfg,
	}

	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		log.Fatalln("create producter failed")
	}

	// 注册信号处理器
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigs
		atomic.AddInt32(&p.running, -1)
		close(p.exitChan)
		p.closeAllConnHandler()
	}()

	p.wg.Wrap(p.timer)

	return p
}

func (p *Producter) OnMessage(f RespMsgFunc) {
	p.event.OnMessage = f
}

func (p *Producter) OnError(f RespMsgFunc) {
	p.event.OnError = f
}

func (p *Producter) Stop() {
	p.wg.Wait()
	log.Println("producer had stop.")
}

// 关闭所有连接处理器
func (p *Producter) closeAllConnHandler() {
	for _, h := range p.handlers {
		h.Do(h.exit)
	}
}

// 定时器,定时轮询注册中心节点,更新节点(新增,删除)
func (p *Producter) timer() {
	defer log.Println("producter timer had exit.")

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-p.exitChan:
			return
		case <-ticker.C:
			nodes, err := GetNodes(p.cfg.RegisterAddr)
			if err != nil {
				log.Println(err)
				continue
			}

			for addr, h := range p.handlers {
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
				if err := p.ConnectToNode(node); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (p *Producter) Publish(j *Job) error {
	if atomic.LoadInt32(&p.running) == 0 {
		return errors.New("Producter is stop")
	}
	if p.weight == 0 {
		return errors.New("Producter.weight is zero")
	}

	if err := j.Validate(); err != nil {
		return err
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Int31n(p.weight)
	intn := int(n) + 1
	w := 0

	for _, hander := range p.handlers {
		nw := w + hander.weight
		if intn >= w && intn <= nw {
			if err := hander.Publish(j); err != nil {
				return err
			} else {
				return nil
			}
		}
		w += nw
	}

	return errors.New("No producter running")
}

// 连接到注册中心
func (p *Producter) ConnectToRegister() error {
	nodes, err := GetNodes(p.cfg.RegisterAddr)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := p.ConnectToNode(node); err != nil {
			return err
		} else {
			atomic.AddInt32(&p.weight, int32(node.Weight))
		}
	}

	return nil
}

// 连接到节点
func (p *Producter) ConnectToNode(node *Node) error {
	if atomic.LoadInt32(&p.running) == 0 {
		return errors.New("producer is waiting to run.")
	}

	conn, err := net.Dial("tcp", node.TcpAddr)
	if err != nil {
		return err
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	if _, ok := p.handlers[node.TcpAddr]; ok {
		return errors.New(fmt.Sprintf("addr %s has exist\n", node.TcpAddr))
	}

	handler := &ProducterConnHander{
		Addr:      node.TcpAddr,
		weight:    node.Weight,
		producter: p,
		conn:      conn,
		pushChan:  make(chan *Job),
		exitChan:  make(chan struct{}),
	}

	p.handlers[node.TcpAddr] = handler
	p.wg.Wrap(handler.Run)

	return nil
}

// 连接到指定节点
func (p *Producter) ConnectSpecifiedNode(tcpAddr string, weight int) error {
	node := &Node{
		TcpAddr: tcpAddr,
		Weight:  weight,
	}

	return p.ConnectToNode(node)
}
