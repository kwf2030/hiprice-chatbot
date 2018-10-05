package times

import (
  "math"
  "math/rand"
  "sync"
  "time"
)

const (
  stateReady   = iota
  stateRunning
  stateStopped
)

var DefaultTimingWheel = NewTimingWheel(60, time.Second)

type TimingWheel struct {
  // 单个slot时间
  duration time.Duration
  ticker   *time.Ticker

  // 当前所在slot
  cur uint64

  // 每个slot对应一个bucket，即一轮共有len(buckets)个slot，
  // 每个bucket是一个map，包含该slot的所有task
  buckets []map[uint64]*task
  // 一轮有多少个slot，等于len(buckets)
  slots uint64

  // 计时器停止信号
  stopCh chan struct{}

  // 用于随机生成task的id
  r *rand.Rand

  l *sync.Mutex

  state int
}

func NewTimingWheel(slots int, duration time.Duration) *TimingWheel {
  arr := make([]map[uint64]*task, slots)
  for i := range arr {
    arr[i] = make(map[uint64]*task, 10)
  }
  return &TimingWheel{
    duration: duration,
    ticker:   time.NewTicker(duration),
    cur:      0,
    buckets:  arr,
    slots:    uint64(slots),
    stopCh:   make(chan struct{}),
    r:        rand.New(rand.NewSource(time.Now().Unix())),
    l:        &sync.Mutex{},
    state:    stateReady,
  }
}

func (tw *TimingWheel) Start() {
  b := false
  tw.l.Lock()
  if tw.state == stateReady {
    tw.state = stateRunning
    b = true
  }
  tw.l.Unlock()
  if b {
    go tw.run()
  }
}

func (tw *TimingWheel) Stop() {
  b := false
  tw.l.Lock()
  if tw.state == stateRunning {
    tw.state = stateStopped
    b = true
  }
  tw.l.Unlock()
  if b {
    close(tw.stopCh)
  }
}

func (tw *TimingWheel) Delay(delay time.Duration, data interface{}, f func(uint64, interface{})) uint64 {
  n1 := int64(delay/tw.duration) / int64(tw.slots)
  n2 := uint64(delay/tw.duration) % tw.slots
  if n2 == 0 {
    n2 = 1
  }
  tw.l.Lock()
  defer tw.l.Unlock()
  n := tw.cur + n2
  task := &task{
    id:    (n << 32) | (tw.r.Uint64() >> 32),
    round: n1,
    data:  data,
    f:     f,
  }
  tw.buckets[n][task.id] = task
  return task.id
}

func (tw *TimingWheel) At(t time.Time, data interface{}, f func(uint64, interface{})) uint64 {
  now := time.Now()
  if t.Before(now) {
    return 0
  }
  return tw.Delay(t.Sub(now), data, f)
}

func (tw *TimingWheel) Cancel(id uint64) {
  i := id >> 32
  if i < tw.slots {
    tw.l.Lock()
    delete(tw.buckets[i], id)
    tw.l.Unlock()
  }
}

func (tw *TimingWheel) run() {
out:
  for {
    select {
    case <-tw.ticker.C:
      tw.l.Lock()
      if tw.cur == tw.slots-1 {
        tw.cur = 0
      } else {
        tw.cur++
      }
      tw.tick()
      tw.l.Unlock()

    case <-tw.stopCh:
      tw.ticker.Stop()
      break out
    }
  }
}

func (tw *TimingWheel) tick() {
  arr := make([]*task, 0, 2)
  for _, t := range tw.buckets[tw.cur] {
    if t.round <= 0 {
      arr = append(arr, t)
    }
    if t.round > math.MinInt64 {
      t.round--
    }
  }
  for _, t := range arr {
    delete(tw.buckets[tw.cur], t.id)
    go t.f(t.id, t.data)
  }
}

type task struct {
  // 高32位用于表示该task所在的slot，低32位随机生成
  id uint64

  // 剩余轮数
  round int64

  data interface{}

  f func(uint64, interface{})
}
