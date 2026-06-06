package id

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	epoch     = int64(1577836800000)
	nodeBits  = uint(10)
	stepBits  = uint(12)
	nodeMax   = int64(-1 ^ (-1 << nodeBits))
	stepMax   = int64(-1 ^ (-1 << stepBits))
	timeShift = nodeBits + stepBits
	nodeShift = stepBits
)

type Snowflake struct {
	mu        sync.Mutex
	epoch     int64
	time      int64
	node      int64
	step      int64
	nodeMax   int64
	stepMax   int64
	timeShift uint
	nodeShift uint
}

func NewSnowflake(node int64) (*Snowflake, error) {
	if node < 0 || node > nodeMax {
		return nil, fmt.Errorf("invalid node id, must be between 0 and %d", nodeMax)
	}

	return &Snowflake{
		epoch:     epoch,
		time:      0,
		node:      node,
		step:      0,
		nodeMax:   nodeMax,
		stepMax:   stepMax,
		timeShift: timeShift,
		nodeShift: nodeShift,
	}, nil
}

func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixNano() / 1000000

	if s.time == now {
		s.step++
		if s.step > s.stepMax {
			for now <= s.time {
				now = time.Now().UnixNano() / 1000000
			}
			s.step = 0
		}
	} else {
		s.step = 0
	}

	s.time = now

	id := int64((now - s.epoch) << s.timeShift)
	id |= int64(s.node << s.nodeShift)
	id |= int64(s.step)

	return id
}

var defaultSnowflake *Snowflake
var once sync.Once

func Init(node int64) error {
	var err error
	once.Do(func() {
		defaultSnowflake, err = NewSnowflake(node)
	})
	return err
}

func Generate() (int64, error) {
	if defaultSnowflake == nil {
		return 0, errors.New("snowflake not initialized, call Init() first")
	}
	return defaultSnowflake.Generate(), nil
}
