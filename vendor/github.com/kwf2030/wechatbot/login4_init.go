package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "net/url"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
)

const initURL = "/webwxinit"

const ContactSelfOp = 0x01

type InitReq struct {
  req *req
}

func (r *InitReq) Run(s *flow.Step) {
  logger.Info().Msg("login, 4th step")
  e := r.validate(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 4th step failed")
    s.Complete(e)
    return
  }
  resp, e := r.do(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 4th step failed")
    s.Complete(e)
    return
  }
  u := conv.Map(resp, "User")
  r.req.userName = conv.String(u, "UserName")
  r.req.avatarURL = fmt.Sprintf("https://%s%s", r.req.host, u["HeadImgUrl"])
  r.req.syncKey = conv.Map(resp, "SyncKey")
  r.req.op <- &op{What: ContactSelfOp, Data: u}
  logger.Info().Msgf("username=%s", r.req.userName)
  s.Complete(nil)
}

func (r *InitReq) validate(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  if r.req.payload == nil {
    return errInvalidArgs
  }
  return nil
}

func (r *InitReq) do(s *flow.Step) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.baseURL + initURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.passTicket)
  q.Set("r", timestampString10())
  addr.RawQuery = q.Encode()
  buf, _ := json.Marshal(r.req.payload)
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
