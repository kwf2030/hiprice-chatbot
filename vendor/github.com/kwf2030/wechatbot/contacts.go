package wechatbot

import (
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/times"
)

const (
  idInitial       = uint64(2018E4)
  idGeneralOffset = uint64(300)
  idGlobalOffset  = uint64(1E7)
  idSelfOffset    = uint64(1E6)
)

// UserName=>ID
// 表示一些内置账号的ID的偏移量
var internalIDs = map[string]uint64{
  "weixin":     1,
  "filehelper": 2,
  "fmessage":   3,
}

type Contacts struct {
  Bot *Bot

  // UserName=>*Contact
  contacts sync.Map

  // ID=>UserName
  userNames sync.Map

  // 当前所有联系人最大的ID，用于生成下一次的联系人ID，
  // 每次启动时会遍历所有ID，找出最大的赋值给maxID，
  // 后面若要再生成ID，用此值自增转为string即可
  maxID uint64

  mu sync.Mutex
}

func initContacts(data []*Contact, bot *Bot) *Contacts {
  logger.Info().Msg("initContacts")
  ret := &Contacts{
    Bot:       bot,
    contacts:  sync.Map{},
    userNames: sync.Map{},
  }
  if len(data) == 0 {
    return ret
  }
  enabled, ok := bot.Attr[AttrPersistentIDEnabled]
  if !ok || !enabled.(bool) {
    for _, c := range data {
      c.Bot = bot
      ret.contacts.Store(c.UserName, c)
    }
    return ret
  }
  // 第一次循环，处理已备注的联系人
  for _, v := range data {
    if v.Flag != ContactFriend {
      continue
    }
    v.Bot = bot
    if remark := conv.String(v.Raw, "RemarkName"); remark != "" {
      if id := parseRemarkToID(remark); id != 0 {
        if ret.maxID < id {
          ret.maxID = id
        }
        v.ID = strconv.FormatUint(id, 10)
        ret.contacts.Store(v.UserName, v)
        ret.userNames.Store(v.ID, v.UserName)
      }
    }
  }
  if ret.maxID == 0 {
    // 如果Bot从未设置过联系人的ID（备注），那么起始ID就是根据当前已经运行的Bot的个数来决定的，
    // 这是为了当有多个Bot的时候，每个Bot的联系人ID是唯一的且不会重复
    l := uint64(CountBots() - 1)
    ret.maxID = idInitial + (l * idGlobalOffset) + idGeneralOffset
  }
  initial := ret.initialID()
  // 第二次循环，处理其他联系人
  for _, v := range data {
    if (v.Flag != ContactFriend && v.Flag != ContactSystem) || v.ID != "" {
      continue
    }
    v.Bot = bot
    if n, ok := internalIDs[v.UserName]; ok {
      v.ID = strconv.FormatUint(initial+n, 10)
    } else if v.UserName == v.Bot.req.userName {
      v.ID = strconv.FormatUint(initial+idSelfOffset, 10)
    } else {
      // 生成一个ID并备注
      v.ID = strconv.FormatUint(ret.NextID(), 10)
      ret.Bot.req.Remark(v.UserName, v.ID)
      time.Sleep(times.RandMillis(times.OneSecondInMillis, times.ThreeSecondsInMillis))
    }
    ret.contacts.Store(v.UserName, v)
    ret.userNames.Store(v.ID, v.UserName)
  }
  // 第三次循环，处理群聊
  for _, v := range data {
    if v.Flag != ContactGroup {
      continue
    }
    v.Bot = bot
    // todo 群没有备注，默认用MaxID自增作为ID，然后用该ID和群名称建立对应关系来解决持久化问题，
    // todo 若群改名，会收到消息，需要在接收消息的地方处理
  }
  // 第四次循环，处理其他类型（ContactMPS等）的联系人，
  // 这类联系人没有ID，只能通过UserName/NickName或关键字索引，
  // 即只有UserName=>*Contact的对应关系，没有ID=>UserName的对应关系
  for _, v := range data {
    if v.ID == "" {
      v.Bot = bot
      ret.contacts.Store(v.UserName, v)
    }
  }
  if bot.Self != nil {
    bot.Self.ID = strconv.FormatUint(initial+idSelfOffset, 10)
  }
  logger.Info().Msg("initContacts, ok")
  return ret
}

func (c *Contacts) Add(contact *Contact) {
  if contact == nil {
    return
  }
  if v, ok := c.contacts.Load(contact.UserName); ok {
    if o, ok := v.(*Contact); ok {
      c.contacts.Delete(o.UserName)
      c.userNames.Delete(o.ID)
      if contact.ID == "" {
        contact.ID = o.ID
      }
    }
  }
  c.contacts.Store(contact.UserName, contact)
  if contact.ID != "" {
    c.userNames.Store(contact.ID, contact.UserName)
  }
}

func (c *Contacts) Remove(userName string) {
  if userName == "" {
    return
  }
  if v, ok := c.contacts.Load(userName); ok {
    c.contacts.Delete(userName)
    if vv, ok := v.(*Contact); ok {
      c.userNames.Delete(vv.ID)
    }
  }
}

func (c *Contacts) Size() int {
  ret := 0
  c.Each(func(_ *Contact) bool {
    ret++
    return true
  })
  return ret
}

func (c *Contacts) FindByID(id string) *Contact {
  if id == "" {
    return nil
  }
  if userName, ok := c.userNames.Load(id); ok {
    if v, ok := c.contacts.Load(userName); ok {
      if ret, ok := v.(*Contact); ok {
        return ret
      }
    }
  }
  return nil
}

func (c *Contacts) FindByUserName(userName string) *Contact {
  if userName == "" {
    return nil
  }
  if v, ok := c.contacts.Load(userName); ok {
    if ret, ok := v.(*Contact); ok {
      return ret
    }
  }
  return nil
}

func (c *Contacts) FindByNickName(nickName string) *Contact {
  if nickName == "" {
    return nil
  }
  var ret *Contact
  c.Each(func(cc *Contact) bool {
    if nickName == cc.Nickname {
      ret = cc
      return false
    }
    return true
  })
  return ret
}

// FindByKeyword根据NickName/拼音查找联系人
func (c *Contacts) FindByKeyword(keyword string) []*Contact {
  if keyword == "" {
    return nil
  }
  ret := make([]*Contact, 0, 1)
  var py1, py2 string
  c.Each(func(cc *Contact) bool {
    py1 = conv.String(cc.Raw, "PYInitial")
    py2 = conv.String(cc.Raw, "PYQuanPin")
    if strings.Contains(cc.Nickname, keyword) {
      ret = append(ret, cc)
      return true
    }
    if py1 != "" && strings.Contains(py1, keyword) {
      ret = append(ret, cc)
      return true
    }
    if py2 != "" && strings.Contains(py2, keyword) {
      ret = append(ret, cc)
      return true
    }
    return true
  })
  return ret
}

// Each遍历所有联系人，action返回false表示终止遍历
func (c *Contacts) Each(action func(cc *Contact) bool) {
  c.contacts.Range(func(k, v interface{}) bool {
    if vv, ok := v.(*Contact); ok {
      return action(vv)
    }
    return true
  })
}

func (c *Contacts) NextID() uint64 {
  c.mu.Lock()
  c.maxID++
  c.mu.Unlock()
  return c.maxID
}

func (c *Contacts) initialID() uint64 {
  if v, ok := c.Bot.Attr[attrInitialID]; ok {
    return v.(uint64)
  }
  str := strconv.FormatUint(c.maxID, 10)
  str = str[:len(str)-4]
  ret, _ := strconv.ParseUint(str, 10, 64)
  ret *= 10000
  c.Bot.Attr[attrInitialID] = ret
  return ret
}
