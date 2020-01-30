package broker

import (
	"github.com/minus5/svckit/amp"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type counter struct {
	msgCount int
	sync.Mutex
}

func (c *counter) Send(m *amp.Msg) {
	c.Lock()
	defer c.Unlock()
	time.Sleep(time.Nanosecond)
	c.msgCount++
}

func TestSpreader(t *testing.T) {
	s := newSpreader("m")
	m1 := &amp.Msg{Ts: 10, UpdateType: amp.Full}
	m2 := &amp.Msg{Ts: 11, UpdateType: amp.Diff}
	m3 := &amp.Msg{Ts: 12, UpdateType: amp.Diff}
	m4 := &amp.Msg{Ts: 13, UpdateType: amp.Diff}
	m5 := &amp.Msg{Ts: 14, UpdateType: amp.Diff}
	m6 := &amp.Msg{Ts: 15, UpdateType: amp.Diff}
	s.publish(m1)
	s.publish(m2)
	s.publish(m4)
	s.publish(m3)
	s.wait()
	msgs := s.replay()
	assert.Len(t, msgs, 4)
	assert.Equal(t, m1.Ts, msgs[0].Ts)
	assert.Equal(t, m2.Ts, msgs[1].Ts)
	assert.Equal(t, m3.Ts, msgs[2].Ts)
	assert.Equal(t, m4.Ts, msgs[3].Ts)
	c1 := counter{}
	c2 := counter{}
	c3 := counter{}
	c4 := counter{}
	c5 := counter{}
	s.subscribe(&c1, 0)
	s.subscribe(&c2, 0)
	s.wait()
	assert.Equal(t, 6, c1.msgCount)
	assert.Equal(t, 6, c2.msgCount)
	s.subscribe(&c3, 0)
	s.subscribe(&c4, 0)
	s.publish(m5)
	s.wait()
	assert.Equal(t, 7, c1.msgCount)
	assert.Equal(t, 7, c2.msgCount)
	empty := s.unsubscribe(&c1)
	assert.False(t, empty)
	s.subscribe(&c5, 0)
	s.publish(m6)
	s.wait()
	assert.Equal(t, 7, c1.msgCount)
	assert.Equal(t, 8, c2.msgCount)
	empty = s.unsubscribe(&c1)
	assert.False(t, empty)
	empty = s.unsubscribe(&c2)
	assert.False(t, empty)
	empty = s.unsubscribe(&c3)
	assert.False(t, empty)
	empty = s.unsubscribe(&c4)
	assert.False(t, empty)
	empty = s.unsubscribe(&c5)
	assert.True(t, empty)
	s.subscribe(&c2, 0)
	s.wait()
	assert.Equal(t, 16, c2.msgCount)
}

type publisher interface {
	subscribe(amp.Subscriber, int64)
	publish(*amp.Msg)
	close()
}

func benchPublisher(p publisher) {
	for i := 0; i < 5000; i++ {
		p.subscribe(&counter{}, 0)
	}
	p.publish(&amp.Msg{Ts: 1, UpdateType: amp.Full})
	for i := 0; i < 5000; i++ {
		p.publish(&amp.Msg{Ts: int64(i + 1), UpdateType: amp.Diff})
	}
	p.close()
}

func BenchmarkTopic(b *testing.B) {
	benchPublisher(newTopic("m"))
}

func BenchmarkSpreader(b *testing.B) {
	benchPublisher(newSpreader("m"))
}
