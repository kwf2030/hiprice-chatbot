package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
  "regexp"
  "strconv"
  "strings"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
)

const (
  syncCheckURL = "/synccheck"
  syncURL      = "/webwxsync"
)

const (
  MsgOp        = 0x20
  ContactModOp = 0x21
  ContactDelOp = 0x22
  TerminateOp  = 0x23
)

var syncCheckRegexp = regexp.MustCompile(`retcode\s*:\s*"(\d+)"\s*,\s*selector\s*:\s*"(\d+)"`)

type SyncReq struct {
  req *req
}

func (r *SyncReq) Run(s *flow.Step) {
  logger.Info().Msg("login, 7th step")
  e := r.validate(s)
  if e != nil {
    logger.Error().Err(e).Msg("login, 7th step failed")
    s.Complete(e)
    return
  }
  // syncCheck一直执行，有消息时才会执行sync，
  // web微信syncCheck的时间间隔约为25秒左右，
  // 即在没有新消息的时候，服务器会保持（阻塞）连接25秒左右
  syncCheck := make(chan struct{})
  sync := make(chan struct{})
  go r.loopSyncCheck(syncCheck, sync)
  go r.loopSync(syncCheck, sync)
  syncCheck <- struct{}{}
  s.Complete(r.req.op)
}

func (r *SyncReq) validate(s *flow.Step) error {
  if e, ok := s.Arg.(error); ok {
    return e
  }
  return nil
}

func (r *SyncReq) loopSyncCheck(syncCheck chan struct{}, sync chan struct{}) {
  for range syncCheck {
    resp, e := r.doSyncCheck()
    code := conv.Int(resp, "code")
    selector := conv.Int(resp, "selector")
    logger.Debug().Msgf("sync check, code=%d, selector=%d", code, selector)
    switch {
    case e != nil, resp == nil:
      fallthrough

    case code == 0 && selector == 0:
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      go func() { syncCheck <- struct{}{} }()
      continue

    case code != 0:
      close(syncCheck)
      close(sync)
      r.req.op <- &op{What: TerminateOp, Data: code}
      close(r.req.op)

    default:
      sync <- struct{}{}
    }
  }
}

func (r *SyncReq) loopSync(syncCheck chan struct{}, sync chan struct{}) {
  for range sync {
    resp, e := r.doSync()
    logger.Debug().Msg("sync")
    switch {
    case e != nil, resp == nil:
      fallthrough

    case conv.Int(conv.Map(resp, "BaseResponse"), "Ret") != 0:
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
      syncCheck <- struct{}{}
      continue
    }

    r.req.syncKey = conv.Map(resp, "SyncCheckKey")

    // 没开启验证如果被添加好友，
    // ModContactList（对方信息）和AddMsgList（添加到通讯录的系统提示）会一起收到，
    // 要先处理完Contact后再处理Message（否则会出现找不到发送者的问题），
    // 虽然之后也能一直收到此人的消息，但要想主动发消息，仍需要手动添加好友，
    // 不添加的话下次登录时好友列表中也没有此人，
    // 目前Web微信好像没有添加好友的功能，所以只能开启验证（通过验证即可添加好友）
    if conv.Int(resp, "ModContactCount") > 0 {
      data := conv.Slice(resp, "ModContactList")
      for _, v := range data {
        r.req.op <- &op{What: ContactModOp, Data: v}
      }
    }
    if conv.Int(resp, "DelContactCount") > 0 {
      data := conv.Slice(resp, "DelContactList")
      for _, v := range data {
        r.req.op <- &op{What: ContactDelOp, Data: v}
      }
    }
    if conv.Int(resp, "AddMsgCount") > 0 {
      data := conv.Slice(resp, "AddMsgList")
      for _, v := range data {
        r.req.op <- &op{What: MsgOp, Data: v}
      }
    }

    time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
    syncCheck <- struct{}{}
  }
}

// 检查是否有新消息，类似于心跳，
// window.synccheck={retcode:"0",selector:"2"}
// retcode=0：正常，
// retcode=1100：失败/已退出，
// retcode=1101：在其他地方登录了Web微信，
// retcode=1102：主动退出，
// selector=0：正常，
// selector=2：有新消息，
// selector=4：保存群聊到通讯录/修改群名称/新增或删除联系人/群聊成员数目变化，
// selector=5：未知，
// selector=6：未知，
// selector=7：操作了手机，如进入/关闭聊天页面
func (r *SyncReq) doSyncCheck() (map[string]interface{}, error) {
  addr, _ := url.Parse(fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin%s", r.req.syncCheckHost, syncCheckURL))
  q := addr.Query()
  q.Set("r", timestampString13())
  q.Set("sid", r.req.sid)
  q.Set("uin", strconv.Itoa(r.req.uin))
  q.Set("skey", r.req.skey)
  q.Set("deviceid", deviceID())
  q.Set("synckey", r.flatSyncKeys())
  q.Set("_", timestampString13())
  addr.RawQuery = q.Encode()
  // 请求必须加上Cookies
  req, _ := http.NewRequest("GET", addr.String(), nil)
  req.Header.Set("Referer", r.req.referer)
  req.Header.Set("User-Agent", userAgent)
  resp, e := r.req.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errReq
  }
  return parseSyncCheckResp(resp)
}

func (r *SyncReq) doSync() (map[string]interface{}, error) {
  addr, _ := url.Parse(r.req.baseURL + syncURL)
  q := addr.Query()
  q.Set("pass_ticket", r.req.passTicket)
  q.Set("sid", r.req.sid)
  q.Set("skey", r.req.skey)
  addr.RawQuery = q.Encode()
  m := r.req.payload
  m["SyncKey"] = r.req.syncKey
  m["rr"] = strconv.FormatInt(^(timestamp() / int64(time.Second)), 10)
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

func (r *SyncReq) flatSyncKeys() string {
  l := conv.Int(r.req.syncKey, "Count")
  list := conv.Slice(r.req.syncKey, "List")
  if len(list) == 0 || len(list) != l {
    return ""
  }
  var sb strings.Builder
  for i := 0; i < l; i++ {
    v := list[i]
    fmt.Fprintf(&sb, "%d_%d", conv.Int(v, "Key"), conv.Int(v, "Val"))
    if i != l-1 {
      sb.WriteString("|")
    }
  }
  return sb.String()
}

func parseSyncCheckResp(resp *http.Response) (map[string]interface{}, error) {
  body, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return nil, e
  }
  data := string(body)
  match := syncCheckRegexp.FindStringSubmatch(data)
  if len(match) < 2 {
    return nil, errResp
  }
  code, _ := strconv.Atoi(match[1])
  selector := 0
  if len(match) >= 3 {
    selector, _ = strconv.Atoi(match[2])
  }
  return map[string]interface{}{"code": code, "selector": selector}, nil
}
