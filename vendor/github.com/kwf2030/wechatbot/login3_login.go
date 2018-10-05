package wechatbot

import (
  "encoding/xml"
  "io/ioutil"
  "net/http"
  "net/url"
  "strings"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
)

type LoginReq struct {
  req *req
}

func (r *LoginReq) Run(s *flow.Step) {
  logger.Info().Msg("login, 3rd step")
  e := r.validate(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 3rd step failed")
    s.Complete(e)
    return
  }
  resp, e := r.do(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 3rd step failed")
    s.Complete(e)
    return
  }
  r.req.payload = map[string]interface{}{
    "BaseRequest": map[string]interface{}{
      "Uin":      resp["wxuin"],
      "Sid":      resp["wxsid"],
      "Skey":     resp["skey"],
      "DeviceID": deviceID(),
    },
  }
  r.req.uin = conv.Int(resp, "wxuin")
  r.req.sid = conv.String(resp, "wxsid")
  r.req.skey = conv.String(resp, "skey")
  r.req.passTicket = conv.String(resp, "pass_ticket")
  r.selectBaseURL(s, r.req.redirectURL)
  logger.Info().Msgf("uin=%d", r.req.uin)
  s.Complete(nil)
}

func (r *LoginReq) validate(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  if r.req.redirectURL == "" {
    return errInvalidArgs
  }
  return nil
}

func (r *LoginReq) do(s *flow.Step) (map[string]interface{}, error) {
  u, _ := url.Parse(r.req.redirectURL)
  // 返回的地址可能没有fun和version两个参数，而此请求必须这两个参数
  q := u.Query()
  q.Set("fun", "new")
  q.Set("version", "v2")
  u.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", u.String(), nil)
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errResp
  }
  return parseLoginResp(resp)
}

func (r *LoginReq) selectBaseURL(s *flow.Step, addr string) {
  u, _ := url.Parse(addr)
  host := u.Hostname()
  r.req.host = host
  switch {
  case strings.Contains(host, "wx2"):
    r.req.referer = "https://wx2.qq.com/"
    r.req.host = "wx2.qq.com"
    r.req.syncCheckHost = "webpush.wx2.qq.com"
    r.req.baseURL = "https://wx2.qq.com/cgi-bin/mmwebwx-bin"
  }
}

func parseLoginResp(resp *http.Response) (map[string]interface{}, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  v := struct {
    XMLName     xml.Name `xml:"error"`
    Ret         int      `xml:"ret"`
    Message     string   `xml:"message"`
    SKey        string   `xml:"skey"`
    WXSid       string   `xml:"wxsid"`
    WXUin       int      `xml:"wxuin"`
    PassTicket  string   `xml:"pass_ticket"`
    IsGrayScale int      `xml:"isgrayscale"`
  }{}
  e = xml.Unmarshal(body, &v)
  if e != nil {
    return nil, e
  }
  return map[string]interface{}{
    "ret":         v.Ret,
    "message":     v.Message,
    "skey":        v.SKey,
    "wxsid":       v.WXSid,
    "wxuin":       v.WXUin,
    "pass_ticket": v.PassTicket,
    "isgrayscale": v.IsGrayScale,
  }, nil
}
