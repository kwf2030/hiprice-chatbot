package wechatbot

import (
  "bytes"
  "encoding/json"
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
)

const statusNotifyURL = "/webwxstatusnotify"

// 在手机上显示"已登录Web微信"
type StatusNotifyReq struct {
  req *req
}

func (r *StatusNotifyReq) Run(s *flow.Step) {
  logger.Info().Msg("login, 5th step")
  e := r.validate(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 5th step failed")
    s.Complete(e)
    return
  }
  _, e = r.do(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 5th step failed")
    s.Complete(e)
    return
  }
  s.Complete(nil)
}

func (r *StatusNotifyReq) validate(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  if r.req.syncKey == nil {
    return errInvalidArgs
  }
  return nil
}

func (r *StatusNotifyReq) do(s *flow.Step) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.baseURL + statusNotifyURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.passTicket)
  addr.RawQuery = q.Encode()
  m := r.req.payload
  m["Code"] = 3
  m["FromUserName"] = r.req.userName
  m["ToUserName"] = r.req.userName
  m["ClientMsgId"] = timestampString13()
  buf, _ := json.Marshal(m)
  // 请求必须加上Content-Type和Cookies
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errReq
  }
  return conv.ReadJSONToMap(resp.Body)
}
