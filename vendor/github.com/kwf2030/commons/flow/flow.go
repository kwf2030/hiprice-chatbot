package flow

import (
  "container/list"
  "errors"
  "math/rand"
  "strconv"
  "sync"
  "time"
)

var (
  ErrCanceled = errors.New("canceled")
  ErrTimeout  = errors.New("timeout")
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

type Flow struct {
  timeout time.Duration
  steps   *list.List

  head *headStep
  tail *tailStep

  finished bool

  result chan interface{}

  sync.Mutex
}

func NewFlow(timeout time.Duration) *Flow {
  f := &Flow{timeout: timeout, steps: list.New(), result: make(chan interface{})}
  f.head = &headStep{Name: strconv.Itoa(rnd.Int()), Flow: f}
  f.head.ele = f.steps.PushFront(f.head)
  f.tail = &tailStep{Name: strconv.Itoa(rnd.Int()), Flow: f}
  f.tail.ele = f.steps.PushBack(f.tail)
  return f
}

func (f *Flow) Start(in interface{}) (interface{}, error) {
  if f.steps.Len() <= 2 {
    return nil, nil
  }
  f.head.inTime = time.Now()
  f.head.Arg = in
  go f.head.Run((*Step)(f.head))
  if f.timeout <= 0 {
    r := <-f.result
    if e, ok := r.(error); ok {
      return nil, e
    }
    return r, nil
  }
  select {
  case r := <-f.result:
    e, ok := r.(error)
    if ok {
      return nil, e
    }
    return r, nil
  case <-time.After(f.timeout):
    f.Lock()
    defer f.Unlock()
    f.finished = true
    f.tail.outTime = time.Now()
    close(f.result)
    return nil, ErrTimeout
  }
  return nil, nil
}

func (f *Flow) Cancel() {
  if f.finished {
    return
  }
  f.Lock()
  defer f.Unlock()
  f.finished = true
  f.tail.outTime = time.Now()
  f.result <- ErrCanceled
  close(f.result)
}

func (f *Flow) Elapsed() time.Duration {
  var t time.Time
  if f.finished {
    t = f.tail.outTime
  } else {
    t = time.Now()
  }
  return t.Sub(f.head.inTime)
}

func (f *Flow) Chain() []string {
  ret := make([]string, 0, f.steps.Len())
  for v := f.head.ele.Next(); v != nil && v != f.tail.ele; v = v.Next() {
    ret = append(ret, v.Value.(*Step).Name)
  }
  return ret
}

func (f *Flow) AddFirst(r StepRunner, name string) {
  f.Lock()
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertAfter(s, f.head.ele)
  f.Unlock()
}

func (f *Flow) AddLast(r StepRunner, name string) {
  f.Lock()
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertBefore(s, f.tail.ele)
  f.Unlock()
}

func (f *Flow) AddBefore(r StepRunner, name, before string) {
  f.Lock()
  defer f.Unlock()
  old := f.step(before)
  if old == nil {
    return
  }
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertBefore(s, old.ele)
}

func (f *Flow) AddAfter(r StepRunner, name, after string) {
  f.Lock()
  defer f.Unlock()
  old := f.step(after)
  if old == nil {
    return
  }
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertAfter(s, old.ele)
}

func (f *Flow) Remove(name string) {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  defer f.Unlock()
  old := f.step(name)
  if old == nil {
    return
  }
  f.steps.Remove(old.ele)
}

func (f *Flow) RemoveFirst() {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  f.steps.Remove(f.head.ele.Next())
  f.Unlock()
}

func (f *Flow) RemoveLast() {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  f.steps.Remove(f.tail.ele.Prev())
  f.Unlock()
}

func (f *Flow) Replace(r StepRunner, name, replace string) {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  defer f.Unlock()
  old := f.step(replace)
  if old == nil {
    return
  }
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertAfter(s, old.ele)
  f.steps.Remove(old.ele)
}

func (f *Flow) ReplaceFirst(r StepRunner, name string) {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  defer f.Unlock()
  f.steps.Remove(f.head.ele.Next())
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertAfter(s, f.head.ele)
}

func (f *Flow) ReplaceLast(r StepRunner, name string) {
  if f.steps.Len() <= 2 {
    return
  }
  f.Lock()
  defer f.Unlock()
  f.steps.Remove(f.tail.ele.Prev())
  s := &Step{Name: name, Flow: f, runner: r}
  s.ele = f.steps.InsertBefore(s, f.tail.ele)
}

func (f *Flow) Get(name string) StepRunner {
  ret := f.step(name)
  if ret == nil {
    return nil
  }
  return ret.runner
}

func (f *Flow) First() StepRunner {
  if f.steps.Len() <= 2 {
    return nil
  }
  return f.head.ele.Next().Value.(*Step).runner
}

func (f *Flow) Last() StepRunner {
  if f.steps.Len() <= 2 {
    return nil
  }
  return f.tail.ele.Prev().Value.(*Step).runner
}

func (f *Flow) step(name string) *Step {
  if f.steps.Len() <= 2 {
    return nil
  }
  for v := f.head.ele.Next(); v != nil && v != f.tail.ele; v = v.Next() {
    s := v.Value.(*Step)
    if s.Name == name {
      return s
    }
  }
  return nil
}
