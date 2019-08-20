package gmq

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type ConsumerConnHandler struct {
	Addr        string
	consumer    *Consumer
	conn        net.Conn
	nextJobChan chan int
	wg          WaitGroupWrapper
	exitChan    chan struct{}
	running     int32
	sync.Once
}

func (h *ConsumerConnHandler) Run() {
	if !atomic.CompareAndSwapInt32(&h.running, 0, 1) {
		log.Fatalln("Start consumer conn handler failed.")
	}

	defer func() {
		h.wg.Wait()
		log.Println(fmt.Sprintf("(%s) ConsumerConnHandler had exit.", h.Addr))
	}()

	h.nextJobChan <- 1
	h.wg.Wrap(h.writeConn)
	h.wg.Wrap(h.readConn)
}

// 关闭连接处理器
func (h *ConsumerConnHandler) exit() {
	defer func() {
		log.Println(fmt.Sprintf("(%s) Gnode conn had been closed.", h.Addr))
	}()

	atomic.AddInt32(&h.running, -1)
	close(h.exitChan)
	h.conn.Close()

	h.consumer.mux.RLock()
	delete(h.consumer.handers, h.Addr)
	h.consumer.mux.RUnlock()
}

// 读处理
func (h *ConsumerConnHandler) readConn() {
	defer func() {
		log.Println(fmt.Sprintf("(%s) Gnode read conn had been closed.", h.Addr))
		h.Do(h.exit)
	}()

	scanner := bufio.NewScanner(h.conn)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) > 4 { // rtypeLen + dataLen
			var rtype, dataLen int16
			binary.Read(bytes.NewReader(data[:2]), binary.BigEndian, &rtype)
			binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &dataLen)
			plen := 4 + int(dataLen)
			if plen <= len(data) {
				return plen, data[:plen], nil
			}
		}
		return
	})

	for scanner.Scan() {
		b := scanner.Bytes()
		rtype := int16(binary.BigEndian.Uint16(b[:2]))
		data := b[4:]

		go h.notify(data, rtype)
	}
}

// 写处理
func (h *ConsumerConnHandler) writeConn() {
	defer func() {
		log.Println(fmt.Sprintf("(%s) Gnode write conn had been closed.", h.Addr))
		h.Do(h.exit)
	}()

	for {
		select {
		case <-h.exitChan:
			return
		case <-h.nextJobChan:
			pkg, err := NewPop(h.consumer.topic)
			if err != nil {
				log.Println("create cmd pkg failed, ", err)
				continue
			}
			if err := h.send(pkg); err != nil {
				log.Println("send failed, ", err)
				return
			}
		}
	}
}

var n = 0

func (h *ConsumerConnHandler) notify(data []byte, rtype int16) {
	h.consumer.limitChan <- 1
	h.nextJobChan <- 1

	ctx := NewContext()
	ctx.SetConsumerConnHandler(h)

	switch rtype {
	case RESP_ERR:
		ctx.SetErr(data)
		h.consumer.event.OnError(ctx)
	case RESP_JOB:
		job := &Job{}
		if err := json.Unmarshal(data, job); err != nil {
			ctx.SetErr([]byte(err.Error()))
			h.consumer.event.OnError(ctx)
		} else {
			ctx.SetJob(job)
			h.consumer.event.OnJob(ctx)
		}
	case RESP_MSG:
		ctx.SetMsg(data)
		h.consumer.event.OnMessage(ctx)
	}

	<-h.consumer.limitChan
}

// 发送命令
func (h *ConsumerConnHandler) send(p *Pkg) error {
	var err error
	err = binary.Write(h.conn, binary.BigEndian, &p.Version)
	err = binary.Write(h.conn, binary.BigEndian, &p.CmdLen)
	err = binary.Write(h.conn, binary.BigEndian, &p.DataLen)
	err = binary.Write(h.conn, binary.BigEndian, &p.Cmd)
	err = binary.Write(h.conn, binary.BigEndian, &p.Data)
	return err
}
