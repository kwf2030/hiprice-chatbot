package wechatbot

import (
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
  "time"

  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const scanStateURL = "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login"

var (
  scanSTCodeRegexp        = regexp.MustCompile(`code\s*=\s*(\d+)\s*;`)
  scanSTRedirectURLRegexp = regexp.MustCompile(`redirect_uri\s*=\s*"(.*)"`)
)

type ScanStateReq struct {
  req *req
}

func (r *ScanStateReq) Run(s *flow.Step) {
  logger.Info().Msg("login, 2nd step")
  e := r.validate(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 2nd step failed")
    s.Complete(e)
    return
  }
  ch := make(chan string)
  go r.check(s, ch)
  redirectURL := <-ch
  close(ch)
  if redirectURL == "" {
    // 如果是空，基本就是超时（一直没有扫描默认设置了2分钟超时），
    // 微信基本不可能返回200状态码的同时返回空redirect_url
    logger.Error().Err(errTimeout).Msg("login, 2nd step failed")
    s.Complete(errTimeout)
    return
  }
  r.req.redirectURL = redirectURL
  logger.Info().Msgf("redirectURL=%s", redirectURL)
  s.Complete(nil)
}

func (r *ScanStateReq) validate(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  if r.req.uuid == "" {
    return errInvalidArgs
  }
  return nil
}

func (r *ScanStateReq) check(s *flow.Step, ch chan<- string) {
  loop := true
  t := time.AfterFunc(time.Minute*2, func() {
    loop = false
    ch <- ""
  })
out:
  for loop {
    // 200（已确认），201（已扫描），408（未扫描）
    code, addr, _ := r.do(s)
    logger.Debug().Msgf("check, code=%d, addr=%s", code, addr)
    switch code {
    case 200:
      r.req.scanState = QRConfirmed
      t.Stop()
      loop = false
      ch <- addr
      break out

    case 201:
      r.req.scanState = QRScanned
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      continue

    case 408:
      r.req.scanState = QRTimeout
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      continue

    default:
      r.req.scanState = QRReady
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      continue
    }
  }
}

func (r *ScanStateReq) do(s *flow.Step) (int, string, error) {
  addr, _ := url.Parse(scanStateURL)
  q := addr.Query()
  q.Set("uuid", r.req.uuid)
  q.Set("tip", "0")
  q.Set("_", timestampString13())
  q.Set("r", timestampString10())
  q.Set("loginicon", "true")
  addr.RawQuery = q.Encode()
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return 0, "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return 0, "", errReq
  }
  // RedirectURL的Host可能是wx.qq.com、wx2.qq.com或其他地址，
  // 这个地址可能是根据帐号注册时间分配的，
  // 从下一步reqToken开始所有的请求必须使用相同的Host，否则会返回1100错误码，
  // wx2版本有些请求的query参数被省略了，暂时不用管
  return parseScanStateResp(resp)
}

func parseScanStateResp(resp *http.Response) (int, string, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return 0, "", e
  }
  // 如果是200，返回：window.code=200;window.redirect_uri=xxx
  // 如果是201，返回：window.code=201;window.userAvatar = 'data:img/jpg;base64,xxx'
  data := string(body)
  code := scanSTCodeRegexp.FindStringSubmatch(data)
  if len(code) != 2 {
    return 0, "", errResp
  }
  c, e := strconv.Atoi(code[1])
  if e != nil {
    return 0, "", errResp
  }
  if c == 200 {
    addr := scanSTRedirectURLRegexp.FindStringSubmatch(data)
    if len(addr) >= 2 {
      return c, addr[1], nil
    }
  }
  return c, "", errResp
}
