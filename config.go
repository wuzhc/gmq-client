package gmq

type ProducerCfg struct {
	RegisterAddr string
}

func NewProducerCfg() *ProducerCfg {
	return &ProducerCfg{
		RegisterAddr: "http://127.0.0.1:9595",
	}
}

type ConsumerCfg struct {
	RegisterAddr  string
	MaxConcurrent int // 并发限制,每次最大MaxConcurrent个消息被处理
}

func NewConsumerCfg() *ConsumerCfg {
	return &ConsumerCfg{
		RegisterAddr:  "http://127.0.0.1:9595",
		MaxConcurrent: 10,
	}
}
