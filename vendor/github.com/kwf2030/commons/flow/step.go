package flow

import (
  "container/list"
  "time"
)

type Step struct {
  ele *list.Element

  runner StepRunner

  inTime  time.Time
  outTime time.Time

  Name string

  Arg interface{}

  Flow *Flow
}

func (s *Step) Complete(out interface{}) {
  t := time.Now()
  s.outTime = t
  next := s.ele.Next().Value
  if v, ok := next.(*tailStep); ok {
    v.inTime = t
    v.Arg = out
    v.Run((*Step)(v))
  } else {
    v := next.(*Step)
    v.inTime = t
    v.Arg = out
    v.runner.Run(v)
  }
}

func (s *Step) AddAfter(r StepRunner, name string) {
  s.Flow.Lock()
  defer s.Flow.Unlock()
  step := &Step{Name: name, Flow: s.Flow, runner: r}
  step.ele = s.Flow.steps.InsertAfter(s, s.ele)
}

func (s *Step) RemoveFollowups() {
  s.Flow.Lock()
  defer s.Flow.Unlock()
  l := make([]*list.Element, 0, 2)
  for v := s.ele.Next(); v != nil && v != s.Flow.tail.ele; v = v.Next() {
    l = append(l, v.Value.(*Step).ele)
  }
  for _, v := range l {
    s.Flow.steps.Remove(v)
  }
}

type headStep Step

func (_ *headStep) Run(s *Step) {
  s.Complete(s.Arg)
}

type tailStep Step

func (_ *tailStep) Run(s *Step) {
  s.Flow.Lock()
  defer s.Flow.Unlock()
  if s.Flow.finished {
    return
  }
  s.Flow.finished = true
  s.outTime = time.Now()
  s.Flow.result <- s.Arg
  close(s.Flow.result)
}

type StepRunner interface {
  Run(s *Step)
}
