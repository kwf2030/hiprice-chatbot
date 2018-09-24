package main

import (
  "encoding/json"
  "time"

  "github.com/kwf2030/commons/beanstalk"
  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/wechatbot"
  "go.etcd.io/bbolt"
)

const attrDisturb = "chatbot.attr.disturb"

func reserveJob(conn *beanstalk.Conn) (string, map[string]interface{}) {
  _, e := conn.Watch(Conf.Beanstalk.ReserveTube)
  if e != nil {
    logger.Error().Err(e).Msg("ERR: Watch")
    return "", nil
  }
  e = conn.Use(Conf.Beanstalk.ReserveTube)
  if e != nil {
    logger.Error().Err(e).Msg("ERR: Use")
    return "", nil
  }
  id, job, e := conn.ReserveWithTimeout(Conf.Beanstalk.ReserveTimeout)
  if e != nil {
    if e != beanstalk.ErrTimedOut {
      logger.Error().Err(e).Msg("ERR: ReserveWithTimeout")
    }
    return "", nil
  }
  ret := make(map[string]interface{})
  e = json.Unmarshal(job, &ret)
  if e != nil {
    logger.Error().Err(e).Msg("ERR: Unmarshal")
    return "", nil
  }
  logger.Info().Msgf("reserve job, ok, job id=%s", id)
  return id, ret
}

func pushLocal() {
  var keys [][]byte
  kv.EachKV(bucketMsgSend, func(k, v []byte, n int) error {
    if len(keys) == 0 {
      keys = make([][]byte, 0, n)
    }
    dst := make([]byte, len(k))
    copy(dst, k)
    keys = append(keys, dst)
    return nil
  })
  if len(keys) == 0 {
    return
  }
  for _, k := range keys {
    if data := kv.Get(bucketMsgSend, k); len(data) > 0 {
      m := make(map[string]interface{}, 4)
      e := json.Unmarshal(data, &m)
      if e != nil {
        continue
      }
      if bu, ok := m["by_user"]; ok {
        if bum, ok := bu.(map[string]interface{}); ok && len(bum) > 0 {
          pushByUser(bum)
        }
      }
      if bt, ok := m["by_text"]; ok {
        if btm, ok := bt.(map[string]interface{}); ok && len(btm) > 0 {
          pushByText(btm)
        }
      }
    }
  }
  kv.UpdateB(bucketMsgSend, func(b *bbolt.Bucket) error {
    for _, k := range keys {
      b.Delete(k)
    }
    return nil
  })
}

// 返回没有推送的消息（开启了免打扰且当前是免打扰时间）
func pushByUser(data map[string]interface{}) map[string]interface{} {
  ret := make(map[string]interface{}, len(data))
  b := isDayTime()
  cnt := 0
  for uid, obj := range data {
    if uid == "" {
      continue
    }
    c := wechatbot.FindContactByID(uid)
    if c == nil {
      continue
    }
    arr, ok := obj.([]interface{})
    if !ok {
      continue
    }
    if !b && !c.GetAttrBool(attrDisturb) {
      // 免打扰时间段且开启了免打扰设置
      ret[uid] = obj
      continue
    }
    for _, v := range arr {
      text, ok := v.(string)
      if !ok || text == "" {
        continue
      }
      if cnt < Conf.Task.MaxSend {
        time.Sleep(times.RandMillis(200, 2000))
      } else {
        time.Sleep(time.Second * time.Duration(Conf.Task.MaxSendDelay))
        cnt = 0
      }
      c.SendText(text)
      cnt++
    }
  }
  return ret
}

// 返回没有推送的消息（开启了免打扰且当前是免打扰时间）
func pushByText(data map[string]interface{}) map[string]interface{} {
  ret := make(map[string]interface{}, len(data))
  b := isDayTime()
  cnt := 0
  for text, obj := range data {
    if text == "" {
      continue
    }
    arr, ok := obj.([]interface{})
    if !ok {
      continue
    }
    users := make([]string, 0, len(arr))
    for _, v := range arr {
      uid, ok := v.(string)
      if !ok || uid == "" {
        continue
      }
      c := wechatbot.FindContactByID(uid)
      if c == nil {
        continue
      }
      if !b && !c.GetAttrBool(attrDisturb) {
        // 免打扰时间段且开启了免打扰设置
        users = append(users, uid)
        continue
      }
      if cnt < Conf.Task.MaxSend {
        time.Sleep(times.RandMillis(200, 2000))
      } else {
        time.Sleep(time.Second * time.Duration(Conf.Task.MaxSendDelay))
        cnt = 0
      }
      c.SendText(text)
      cnt++
    }
    if len(users) > 0 {
      ret[text] = users
    }
  }
  return ret
}
