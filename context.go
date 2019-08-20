package gmq

import (
	"context"
)

type Context struct {
	ctx context.Context
}

func NewContext() *Context {
	ctx := context.Background()
	return &Context{
		ctx: ctx,
	}
}

func (c *Context) SetJob(job *Job) {
	c.ctx = context.WithValue(c.ctx, "job", job)
}

func (c *Context) SetMsg(msg []byte) {
	c.ctx = context.WithValue(c.ctx, "msg", msg)
}

func (c *Context) SetErr(err []byte) {
	c.ctx = context.WithValue(c.ctx, "err", err)
}

func (c *Context) SetConsumerConnHandler(handler *ConsumerConnHandler) {
	c.ctx = context.WithValue(c.ctx, "handler", handler)
}

func (c *Context) SetProducerConnHandler(handler *ProducterConnHander) {
	c.ctx = context.WithValue(c.ctx, "handler", handler)
}

func (c *Context) GetJob() *Job {
	j := c.ctx.Value("job")
	if j == nil {
		return nil
	}

	return j.(*Job)
}

func (c *Context) GetMsg() string {
	m := c.ctx.Value("msg")
	if m == nil {
		return ""
	}

	return string(m.([]byte))
}

func (c *Context) GetErr() string {
	m := c.ctx.Value("err")
	if m == nil {
		return ""
	}

	return string(m.([]byte))
}

func (c *Context) GetConsumerConnHandler() *ConsumerConnHandler {
	v := c.ctx.Value("handler")
	if v == nil {
		return nil
	}

	return v.(*ConsumerConnHandler)
}

func (c *Context) GetProducterConnHander() *ProducterConnHander {
	v := c.ctx.Value("handler")
	if v == nil {
		return nil
	}

	return v.(*ProducterConnHander)
}
