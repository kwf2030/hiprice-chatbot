package main

import (
  "crypto/md5"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "math/rand"
  "regexp"
  "strings"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/httputil"
  "github.com/kwf2030/commons/times"
  "github.com/kwf2030/wechatbot"
)

const (
  attrImage1 = "chatbot.attr.welcome_image1"
  attrImage2 = "chatbot.attr.welcome_image2"
  attrVideo  = "chatbot.attr.welcome_video"
)

var replyTpl = []string{
  "已记录",
  "没问题",
  "收到",
  "明白",
  "了解",
  "知道",
  "好的",
  "OK",
  "Get",
  "Done",
  "Roger That",
  "Copy That",
  "YES, My Lord",
  "No Problem",
  "[OK]",
  "[嘿哈]",
  "[皱眉]",
  "[机智]",
}

var urlRegex = regexp.MustCompile(`[-a-zA-Z0-9@:%_+.~#?&/=]{2,256}\.[a-z]{2,4}\b(/[-a-zA-Z0-9@:%_+.~#?&/=]*)?`)

type dispatcher struct {
  bot        *wechatbot.Bot
  opNotifier <-chan *wechatbot.Op
}

func (dp *dispatcher) loop() {
  for op := range dp.opNotifier {
    go dp.dispatch(op)
  }
}

func (dp *dispatcher) dispatch(op *wechatbot.Op) {
  switch op.What {
  case wechatbot.MsgOp:
    dp.processMsg(op)

  case wechatbot.ContactListOp:
    if dp.bot.GetAttrBool(wechatbot.AttrPersistentIDEnabled) {
      dp.persistContacts()
    }

  case wechatbot.TerminateOp:
    n := dp.bot.Self.Nickname
    t := dp.bot.StopTime.Sub(dp.bot.StartTime)
    logger.Info().Msgf("wechatbot offline, continuous online for %.2f hours", t.Hours())
    dp.bot.Release()
    if Conf.MMS.Enabled != 0 {
      str := fmt.Sprintf("%d小时%d分钟", int(t.Hours()), int(t.Minutes())%60)
      sendOfflineSms(n, str)
    }
  }
}

func (dp *dispatcher) persistContacts() {
  tx, _ := db.Begin()
  defer tx.Commit()
  aid := 0
  dp.bot.Contacts.Each(func(c *wechatbot.Contact) bool {
    disturb := 0
    if c.ID != "" {
      tx.QueryRow(`SELECT _id, disturb FROM user WHERE id=? LIMIT 1`, c.ID).Scan(&aid, &disturb)
      c.Raw[attrDisturb] = disturb != 0
      if aid == 0 {
        tx.Exec(`INSERT INTO user (id, nickname, create_time, uin) VALUES (?, ?, ?, ?)`, c.ID, c.Nickname, times.NowStr(), c.Uin)
      } else {
        tx.Exec(`UPDATE user SET nickname=? WHERE id=?`, c.Nickname, c.ID)
      }
    }
    return true
  })
}

func persistContact(c *wechatbot.Contact) {
  if c.ID == "" {
    return
  }
  tx, _ := db.Begin()
  defer tx.Commit()
  aid := 0
  tx.QueryRow(`SELECT _id FROM user WHERE id=? LIMIT 1`, c.ID).Scan(&aid)
  if aid == 0 {
    tx.Exec(`INSERT INTO user (id, nickname, create_time, uin) VALUES (?, ?, ?, ?)`, c.ID, c.Nickname, times.NowStr(), c.Uin)
  } else {
    tx.Exec(`UPDATE user SET nickname=? WHERE id=?`, c.Nickname, c.ID)
  }
}

func (dp *dispatcher) processMsg(op *wechatbot.Op) {
  logger.Debug().Msg("msg received: " + op.Msg.ID)
  raw, _ := json.Marshal(op.Msg.Raw)
  _, e := db.Exec(`INSERT INTO msg (id, from_user_id, to_user_id, type, content, url, create_time, raw) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
    op.Msg.ID, op.Msg.FromUserID, op.Msg.ToUserID, op.Msg.Type, op.Msg.Content,
    op.Msg.URL, op.Msg.CreateTime.Format(times.DateTimeSFormat), raw)
  if e != nil {
    logger.Error().Err(e).Msg("Err: INSERT")
  }

  // 开启朋友验证才会收到此类消息
  if op.Msg.Type == wechatbot.MsgVerify {
    logger.Debug().Msg("verify user")
    info := conv.Map(op.Msg.Raw, "RecommendInfo")
    id, _ := dp.bot.VerifyAndRemark(conv.String(info, "UserName"), conv.String(info, "Ticket"))
    if id == "" {
      return
    }
    c := wechatbot.FindContactByID(id)
    if c == nil {
      return
    }

    if dp.bot.GetAttrBool(wechatbot.AttrPersistentIDEnabled) {
      persistContact(c)
    }

    // 首次欢迎消息
    c.SendText("把喜欢的宝贝发给我，我会24小时监控它的价格，有波动第一时间通知你")
    time.Sleep(time.Second)

    if mediaID, ok := dp.bot.Attr[attrImage1]; ok {
      dp.bot.ForwardImageToUserName(c.UserName, mediaID.(string))
      time.Sleep(time.Second)
    } else {
      data, e := ioutil.ReadFile("welcome1.png")
      if e == nil {
        mediaID, e := c.SendImage(data, "welcome1.png")
        if e == nil {
          dp.bot.Attr[attrImage1] = mediaID
        }
        time.Sleep(time.Second)
      }
    }
    if mediaID, ok := dp.bot.Attr[attrImage2]; ok {
      dp.bot.ForwardImageToUserName(c.UserName, mediaID.(string))
      time.Sleep(time.Second)
    } else {
      data, e := ioutil.ReadFile("welcome2.png")
      if e == nil {
        mediaID, e := c.SendImage(data, "welcome2.png")
        if e == nil {
          dp.bot.Attr[attrImage2] = mediaID
        }
        time.Sleep(time.Second)
      }
    }

    if mediaID, ok := dp.bot.Attr[attrVideo]; ok {
      dp.bot.ForwardVideoToUserName(c.UserName, mediaID.(string))
      time.Sleep(time.Second)
    } else {
      data, e := ioutil.ReadFile("welcome.mp4")
      if e == nil {
        mediaID, e := c.SendVideo(data, "welcome.mp4")
        if e == nil {
          dp.bot.Attr[attrVideo] = mediaID
        }
      }
    }
    return
  }

  if op.Msg.Type == wechatbot.MsgSystem || op.Msg.Type == wechatbot.MsgInit {
    return
  }

  // 如果不是好友发来的聊天消息不处理（群消息也暂时不处理）
  c := op.Msg.GetFromContact()
  if c == nil || c.Flag != wechatbot.ContactFriend {
    return
  }

  // intercept返回需要回复的文本和是否拦截此消息，
  // 若拦截，回复文本不为空则回复，
  // 若不拦截，则进行下一步处理
  text, ok := intercept(op.Msg)
  if ok {
    if text != "" {
      op.Msg.ReplyText(text)
    }
    return
  }

  if op.Msg.Type == wechatbot.MsgText {
    if op.Msg.Content == "帮助" {
      var addr string
      v := kv.Get(bucketVar, []byte("help"))
      if v == nil {
        s1 := fmt.Sprintf("%s/help", Conf.Server.Web)
        s2 := httputil.ShortenURL(s1)
        if s2 == "" {
          s2 = s1
        } else {
          kv.UpdateV(bucketVar, []byte("help"), []byte(s2))
        }
        addr = s2
      } else {
        addr = string(v)
      }
      op.Msg.ReplyText(fmt.Sprintf("更多好玩的功能\n%s\n随时更新，常来看看哦", addr))
      return
    }
    if op.Msg.Content == "我" {
      v := []byte(op.Msg.FromUserID)
      k := fmt.Sprintf("%x", md5.Sum(v))
      kv.UpdateV(bucketUserID, []byte(k), v)
      s1 := fmt.Sprintf("%s/watchlist?u=%s", Conf.Server.Web, k)
      s2 := httputil.ShortenURL(s1)
      if s2 == "" {
        s2 = s1
      }
      op.Msg.ReplyText("管理关注的宝贝，请移步\n" + s2)
      return
    }

    // 如果检测到是淘宝/天猫APP分享且没有地址的，
    // 提示分享的时候使用复制链接而不是分享到微信
    if strings.Contains(op.Msg.Content, "手淘") && strings.Contains(op.Msg.Content, "复制这条信息") {
      if !urlRegex.MatchString(op.Msg.Content) {
        op.Msg.ReplyText("亲~由于手淘的限制，在手淘APP中点击分享后请选择『复制链接』，再到微信中粘贴发给我，劳烦重发一次喽")
      }
    } else if strings.Contains(op.Msg.Content, "天猫") && strings.Contains(op.Msg.Content, "复制整段信息") && strings.Contains(op.Msg.Content, "喵口令") {
      if !urlRegex.MatchString(op.Msg.Content) {
        op.Msg.ReplyText("亲~由于天猫的限制，在天猫APP中点击分享后请选择『复制口令』或『复制链接』，再到微信中粘贴发给我，劳烦重发一次喽")
      }
    }
  }
  op.Msg.ReplyText(replyTpl[rand.Intn(len(replyTpl))])
}

func intercept(msg *wechatbot.Message) (string, bool) {
  switch {
  case strings.HasPrefix(msg.Content, "#jy#"):
    _, e := db.Exec(`INSERT INTO suggestion (msg_id, from_user_id, content) VALUES (?, ?, ?)`, msg.ID, msg.FromUserID, msg.Content[4:])
    if e != nil {
      logger.Error().Err(e).Msg("Err: INSERT")
      return "", true
    }
    return "建议已收到，非常感谢\n还有什么好想法随时告诉我", true

  case msg.Content == "#lx#":
    return Conf.Email, true
  }

  return "", false
}

// todo 调用阿里云消息服务发送下线短信通知
func sendOfflineSms(name, duration string) {

}
