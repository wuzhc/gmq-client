package gmq

const (
	RESP_JOB = iota // 响应job,用于客户端消费job
	RESP_ERR        // 响应错误
	RESP_MSG        // 响应消息(提示)
)

type Event struct {
	OnJob     RespMsgFunc
	OnError   RespMsgFunc
	OnMessage RespMsgFunc
}

type RespMsgFunc func(ctx *Context)
