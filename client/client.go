package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrTopicEmpty = errors.New("topic is empty")
)

type MsgPkg struct {
	Body  []byte
	Topic string
	Delay string
}

type MMsgPkg struct {
	Body  []byte
	Delay int
}

type Client struct {
	conn   net.Conn
	addr   string
	weight int
}

func NewClient(addr string, weight int) *Client {
	if len(addr) == 0 {
		log.Fatalln("address is empty")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln(err)
	}

	return &Client{
		conn:   conn,
		addr:   addr,
		weight: weight,
	}
}

func (c *Client) Exit() {
	c.conn.Close()
}

// 获取客户端连接的节点地址
func (c *Client) GetAddr() string {
	return c.addr
}

// 消费消息
func (c *Client) Pop(topic string) error {
	if len(topic) == 0 {
		return ErrTopicEmpty
	}

	var params [][]byte
	params = append(params, []byte("pop"))
	params = append(params, []byte(topic))
	line := bytes.Join(params, []byte(" "))
	if _, err := c.conn.Write(line); err != nil {
		return err
	}
	if _, err := c.conn.Write([]byte{'\n'}); err != nil {
		return err
	}

	return nil
}

// 生产消息
func (c *Client) Push(pkg MsgPkg) error {
	if len(pkg.Topic) == 0 {
		return ErrTopicEmpty
	}

	var params [][]byte
	params = append(params, []byte("pub"))
	params = append(params, []byte(pkg.Topic))
	params = append(params, []byte(pkg.Delay))
	line := bytes.Join(params, []byte(" "))
	if _, err := c.conn.Write(line); err != nil {
		return err
	}
	if _, err := c.conn.Write([]byte{'\n'}); err != nil {
		return err
	}

	// write msg.body
	bodylen := make([]byte, 4)
	body := pkg.Body
	binary.BigEndian.PutUint32(bodylen, uint32(len(body)))
	if _, err := c.conn.Write(bodylen); err != nil {
		return err
	}
	if _, err := c.conn.Write(body); err != nil {
		return err
	}

	return nil
}

// 批量生产消息
func (c *Client) Mpush(topic string, msgs []MMsgPkg) error {
	if len(topic) == 0 {
		return ErrTopicEmpty
	}
	lmsg := len(msgs)
	if lmsg == 0 {
		return errors.New("msgs is empty")
	}

	var params [][]byte
	params = append(params, []byte("mpub"))
	params = append(params, []byte(topic))
	params = append(params, []byte(strconv.Itoa(lmsg)))
	line := bytes.Join(params, []byte(" "))
	if _, err := c.conn.Write(line); err != nil {
		return err
	}
	if _, err := c.conn.Write([]byte{'\n'}); err != nil {
		return err
	}

	for i := 0; i < lmsg; i++ {
		delay := make([]byte, 4)
		binary.BigEndian.PutUint32(delay, uint32(msgs[i].Delay))
		bodylen := make([]byte, 4)
		binary.BigEndian.PutUint32(bodylen, uint32(len(msgs[i].Body)))
		if _, err := c.conn.Write(delay); err != nil {
			return err
		}
		if _, err := c.conn.Write(bodylen); err != nil {
			return err
		}
		if _, err := c.conn.Write(msgs[i].Body); err != nil {
			return err
		}
	}

	return nil
}

// 确认已消费消息
func (c *Client) Ack(msgId string, topic string) error {
	if len(topic) == 0 {
		return ErrTopicEmpty
	}

	var params [][]byte
	params = append(params, []byte("ack"))
	params = append(params, []byte(msgId))
	params = append(params, []byte(topic))
	line := bytes.Join(params, []byte(" "))
	if _, err := c.conn.Write(line); err != nil {
		return err
	}
	if _, err := c.conn.Write([]byte{'\n'}); err != nil {
		return err
	}

	return nil
}

func (c *Client) Recv() (int, []byte) {
	respTypeBuf := make([]byte, 2)
	io.ReadFull(c.conn, respTypeBuf)
	respType := binary.BigEndian.Uint16(respTypeBuf)

	bodyLenBuf := make([]byte, 4)
	io.ReadFull(c.conn, bodyLenBuf)
	bodyLen := binary.BigEndian.Uint32(bodyLenBuf)

	bodyBuf := make([]byte, bodyLen)
	io.ReadFull(c.conn, bodyBuf)

	return int(respType), bodyBuf
}

// 生产消息
func Example_Produce(c *Client, topic string) {
	for i := 0; i < 2; i++ {
		// push message
		msg := MsgPkg{}
		msg.Body = []byte("golang_" + strconv.Itoa(i))
		msg.Topic = topic
		msg.Delay = "0"
		if err := c.Push(msg); err != nil {
			log.Fatalln(err)
		}

		// receive response
		rtype, data := c.Recv()
		log.Println(fmt.Sprintf("rtype:%v, msg.id:%v", rtype, string(data)))
	}
}

// 消费消息
func Example_Consume(c *Client, topic string) {
	if err := c.Pop(topic); err != nil {
		log.Println(err)
	}

	// receive response
	rtype, data := c.Recv()
	if rtype == 0 {
		log.Println(fmt.Sprintf("recv msg:%v", string(data)))
	} else {
		log.Println(string(data))
	}
}

// 轮询模式消费消息
// 当server端没有消息后,sleep 3秒后再次请求
func Example_Loop_Consume(c *Client, topic string) {
	for {
		if err := c.Pop(topic); err != nil {
			log.Println(err)
		}

		// receive response
		rtype, data := c.Recv()
		if rtype == 0 {
			log.Println(fmt.Sprintf("recv msg:%v", string(data)))
		} else {
			log.Println(string(data))
			time.Sleep(3 * time.Second)
		}
	}
}

// 批量生产消息
func Example_MProduce(c *Client, topic string) {
	total := 10
	var msgs []MMsgPkg
	for i := 0; i < total; i++ {
		msgs = append(msgs, MMsgPkg{[]byte("golang_" + strconv.Itoa(i)), 0})
	}
	if err := c.Mpush(topic, msgs); err != nil {
		log.Fatalln(err)
	}

	// receive response
	rtype, data := c.Recv()
	log.Println(fmt.Sprintf("rtype:%v, result:%v", rtype, string(data)))
}

var clients []*Client

// 初始化客户端,建立和注册中心节点连接
func InitClients(registerAddr string) ([]*Client, error) {
	url := fmt.Sprintf("%s/getNodes", registerAddr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Node struct {
		Id       int64  `json:"id"`
		HttpAddr string `json:"http_addr"`
		TcpAddr  string `json:"tcp_addr"`
		JoinTime string `json:"join_time"`
		Weight   int    `json:"weight"`
	}
	v := struct {
		Code int `json:"code"`
		Data struct {
			Nodes []*Node `json:"nodes"`
		} `json:"data"`
		Msg string `json:"msg"`
	}{}

	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	var clients []*Client
	for i := 0; i < len(v.Data.Nodes); i++ {
		c := NewClient(v.Data.Nodes[i].TcpAddr, v.Data.Nodes[i].Weight)
		clients = append(clients, c)
	}

	return clients, nil
}

// 权重模式
func GetClientByWeightMode() *Client {
	if len(clients) == 0 {
		var err error
		clients, err = InitClients("http://127.0.0.1:9595")
		if err != nil {
			log.Fatalln(err)
		}
	}

	total := 0
	for _, c := range clients {
		total += c.weight
	}

	w := 0
	rand.Seed(time.Now().UnixNano())
	randValue := rand.Intn(total) + 1
	for _, c := range clients {
		prev := w
		w = w + c.weight
		if randValue > prev && randValue <= w {
			return c
		}
	}

	return nil
}

// 随机模式
func GetClientByRandomMode() *Client {
	if len(clients) == 0 {
		var err error
		clients, err = InitClients("http://127.0.0.1:9595")
		if err != nil {
			log.Fatalln(err)
		}
	}

	rand.Seed(time.Now().UnixNano())
	k := rand.Intn(len(clients))
	return clients[k]
}

// 平均模式
func GetClientByAvgMode() *Client {
	if len(clients) == 0 {
		var err error
		clients, err = InitClients("http://127.0.0.1:9595")
		if err != nil {
			log.Fatalln(err)
		}
	}

	c := clients[0]
	if len(clients) > 1 {
		// 已处理过的消息客户端重新放在最后
		clients = append(clients[1:], c)
	}

	return c
}
