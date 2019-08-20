package gmq

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type ProducterConnHander struct {
	Addr          string
	weight        int
	producter     *Producter
	conn          net.Conn
	wg            WaitGroupWrapper
	pushChan      chan *Job
	exitChan      chan struct{}
	waitDoneQueue []*Job
	running       int32
	sync.Once
	sync.Mutex
}

func (h *ProducterConnHander) Run() {
	if !atomic.CompareAndSwapInt32(&h.running, 0, 1) {
		log.Fatalln("Start consumer conn handler failed.")
	}

	defer func() {
		h.wg.Wait()
		log.Println(fmt.Sprintf("(%s) ProducterConnHander had exit.", h.Addr))
	}()

	h.wg.Wrap(h.writeConn)
	h.wg.Wrap(h.readConn)
}

// 关闭连接处理器
func (h *ProducterConnHander) exit() {
	defer func() {
		log.Println(fmt.Sprintf("(%s) Gnode conn had been closed.", h.Addr))
	}()

	atomic.AddInt32(&h.running, -1)
	close(h.exitChan)
	h.conn.Close()

	h.producter.mux.RLock()
	delete(h.producter.handlers, h.Addr)
	h.producter.mux.RUnlock()

}

func (h *ProducterConnHander) writeConn() {
	defer func() {
		log.Println(fmt.Sprintf("(%s) Gnode write conn had been closed.", h.Addr))
		h.Do(h.exit)
	}()

	for {
		select {
		case <-h.exitChan:
			return
		case j := <-h.pushChan:
			h.pushWaitQueue(j)

			pkg, err := NewPush(j)
			if err != nil {
				h.popWaitQueue()
				h.notify([]byte(err.Error()), RESP_ERR)
				continue
			}
			if err := h.Send(pkg); err != nil {
				h.popWaitQueue()
				h.notify([]byte(err.Error()), RESP_ERR)
				return
			}
		}
	}
}

func (h *ProducterConnHander) readConn() {
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
		h.notify(data, rtype)
	}
}

func (h *ProducterConnHander) notify(data []byte, rtype int16) {
	job := h.popWaitQueue()
	if job == nil {
		log.Fatalln("System failed...")
	}

	ctx := NewContext()
	ctx.SetProducerConnHandler(h)
	ctx.SetJob(job)

	switch rtype {
	case RESP_ERR:
		ctx.SetErr(data)
		h.producter.event.OnError(ctx)
	case RESP_MSG:
		ctx.SetMsg(data)
		h.producter.event.OnMessage(ctx)
	}
}

func (h *ProducterConnHander) pushWaitQueue(j *Job) {
	h.Lock()
	defer h.Unlock()

	h.waitDoneQueue = append(h.waitDoneQueue, j)
}

func (h *ProducterConnHander) popWaitQueue() *Job {
	h.Lock()
	defer h.Unlock()

	var job *Job
	if len(h.waitDoneQueue) > 0 {
		job = h.waitDoneQueue[0]
		h.waitDoneQueue = h.waitDoneQueue[1:]
	}

	return job
}

func (h *ProducterConnHander) Send(p *Pkg) error {
	var err error
	err = binary.Write(h.conn, binary.BigEndian, &p.Version)
	err = binary.Write(h.conn, binary.BigEndian, &p.CmdLen)
	err = binary.Write(h.conn, binary.BigEndian, &p.DataLen)
	err = binary.Write(h.conn, binary.BigEndian, &p.Cmd)
	err = binary.Write(h.conn, binary.BigEndian, &p.Data)
	return err
}

func (h *ProducterConnHander) Publish(j *Job) error {
	select {
	case h.pushChan <- j:
		return nil
	case <-h.exitChan:
		return errors.New("Conn is closed.")
	case <-time.After(5 * time.Second):
		return errors.New("Publish timeout.")
	}
}
