package main

import (
  "net/http"
  "strconv"
  "sync"
  "time"

  "github.com/kwf2030/wechatbot"
)

var botsHandler = func(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }
  arr := make([]*Bot, 0, wechatbot.CountBots())
  for _, v := range wechatbot.AllBots() {
    if v.State != wechatbot.BotRunning {
      continue
    }
    arr = append(arr, &Bot{v.Self.Uin, v.Self.Nickname, v.GetAttrString(wechatbot.AttrPathAvatar), v.StartTimeStr})
  }
  sendResp(w, 0, "", map[string]interface{}{"bots": arr})
}

var botHandler = func(w http.ResponseWriter, r *http.Request) {
  q := r.URL.Query()
  uin := q.Get("uin")
  uuid := q.Get("uuid")
  switch r.Method {
  case http.MethodGet:
    if uuid != "" {
      getLoginState(w, uuid)
    } else {
      w.WriteHeader(http.StatusBadRequest)
    }

  case http.MethodPost:
    launchBot(w)

  case http.MethodDelete:
    if uin != "" {
      exitBot(w, uin)
    } else {
      w.WriteHeader(http.StatusBadRequest)
    }

  default:
    w.WriteHeader(http.StatusMethodNotAllowed)
  }
}

type Bot struct {
  Uin       int    `json:"uin,omitempty"`
  NickName  string `json:"nickname,omitempty"`
  Avatar    string `json:"avatar,omitempty"`
  StartTime string `json:"start_time,omitempty"`
}

func getLoginState(w http.ResponseWriter, uuid string) {
  b := wechatbot.FindBotByUUID(uuid)
  if b == nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  m := make(map[string]interface{}, 2)
  m["state"] = b.GetScanState()
  if b.State == wechatbot.BotRunning {
    // 如果已经是Running状态，下载头像一起返回
    p, e := b.DownloadAvatar(b.GetAttrString(wechatbot.AttrPathAvatar))
    if e == nil {
      m["bot"] = Bot{b.Self.Uin, b.Self.Nickname, p, b.StartTimeStr}
    }
  }
  sendResp(w, 0, "", m)
}

func launchBot(w http.ResponseWriter) {
  var mu sync.Mutex
  b := true
  time.AfterFunc(time.Minute*2, func() {
    mu.Lock()
    defer mu.Unlock()
    if b {
      b = false
      w.WriteHeader(http.StatusBadRequest)
    }
  })
  ch := make(chan string)
  go func() {
    b := wechatbot.CreateBot(true)
    op, e := b.Start(ch)
    if e != nil || op == nil {
      return
    }
    dp := &dispatcher{b, op}
    go dp.loop()
  }()
  qr := <-ch
  mu.Lock()
  defer mu.Unlock()
  if !b || qr == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  b = false
  sendResp(w, 0, "", map[string]interface{}{"qrcode": qr})
}

func exitBot(w http.ResponseWriter, uin string) {
  u, e := strconv.Atoi(uin)
  if e != nil {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  b := wechatbot.FindBotByUin(u)
  if b == nil {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  b.Stop()
  sendResp(w, 0, "", nil)
}
