package channel

import (
	"sync"

	"github.com/silvernodes/silvernode-go/process"
	"github.com/silvernodes/silvernode-go/utils/snowflake"
)

type msgchannel struct {
	name        string
	subscribers map[int64]*subscriber
	sync.RWMutex
}

func cochannel(name string) *msgchannel {
	c := new(msgchannel)
	c.name = name
	c.subscribers = make(map[int64]*subscriber)
	return c
}

func (c *msgchannel) Name() string {
	return c.name
}

func (c *msgchannel) Publish(datas interface{}) {
	c.RLock()
	defer c.RUnlock()

	for _, sub := range c.subscribers {
		sub.publish(datas)
	}
}

func (c *msgchannel) Subscribe(fun func(interface{}), processor process.Processor) int64 {
	c.Lock()
	defer c.Unlock()

	seq := snowflake.GenerateRaw()
	c.subscribers[seq] = cosubscriber(seq, fun, processor)
	return seq
}

func (c *msgchannel) Unsubscribe(seq int64) {
	c.Lock()
	defer c.Unlock()

	delete(c.subscribers, seq)
}

func (c *msgchannel) UnsubscribeAll() {
	c.Lock()
	defer c.Unlock()

	c.subscribers = nil
	c.subscribers = make(map[int64]*subscriber)
}

func (c *msgchannel) NumOfSubscriber() int {
	c.Lock()
	defer c.Unlock()

	return len(c.subscribers)
}
