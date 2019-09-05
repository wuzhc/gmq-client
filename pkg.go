// 数据包定义,包装,用于网络请求
package gmq

import (
	"encoding/json"
	"errors"
)

const (
	CMD_POP  = "pop"
	CMD_PUSH = "push"
	CMD_PING = "ping"
	CMD_ACK  = "ack"
)

var (
	PKG_VERSION = [4]byte{'v', '1', '1', '1'}
)

type Pkg struct {
	Version [4]byte
	CmdLen  int16
	DataLen int16
	Cmd     []byte
	Data    []byte
}

type Job struct {
	Id         int64  `redis:"id"`
	Topic      string `redis:"topic"`
	Delay      int    `redis:"delay"`
	TTR        int    `redis:"TTR"` // time-to-run
	Body       string `redis:"body"`
	Status     int    `redis:"status"`
	ConsumeNum int    `redis:"consume_num"`
}

func (j *Job) Validate() error {
	if len(j.Topic) == 0 {
		return errors.New("job.topic is empty")
	}
	if len(j.Body) == 0 {
		return errors.New("job.body is empty")
	}
	return nil
}

// 消费请求数据包
func NewPop(topic string) (*Pkg, error) {
	if len(topic) == 0 {
		return nil, errors.New("topic is empty")
	}

	return &Pkg{
		Version: PKG_VERSION,
		Cmd:     []byte(CMD_POP),
		CmdLen:  int16(len(CMD_POP)),
		Data:    []byte(topic),
		DataLen: int16(len(topic)),
	}, nil
}

// 心跳请求数据包
func NewPing() *Pkg {
	return &Pkg{
		Version: PKG_VERSION,
		Cmd:     []byte(CMD_PING),
		CmdLen:  int16(len(CMD_PING)),
	}
}

// 推送请求数据包
func NewPush(j *Job) (*Pkg, error) {
	v, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return &Pkg{
		Version: PKG_VERSION,
		Cmd:     []byte(CMD_PUSH),
		CmdLen:  int16(len(CMD_PUSH)),
		Data:    v,
		DataLen: int16(len(v)),
	}, nil
}
