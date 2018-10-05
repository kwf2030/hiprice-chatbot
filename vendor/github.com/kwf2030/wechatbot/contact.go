package wechatbot

import (
  "strconv"
  "time"

  "github.com/kwf2030/commons/conv"
)

const (
  ContactUnknown = iota
  ContactFriend
  ContactGroup
  ContactMPS
  ContactSystem
)

type Contact struct {
  ID         string    `json:"id,omitempty"`
  Nickname   string    `json:"nickname,omitempty"`
  CreateTime time.Time `json:"create_time,omitempty"`

  // 不是联系人的UIN，是联系人所属Bot的UIN
  Uin int `json:"uin"`

  Raw map[string]interface{} `json:"raw,omitempty"`

  // VerifyFlag表示联系人类型，
  // 个人和群聊帐号为0，
  // 订阅号为8，
  // 企业号为24（包括扩微信支付），
  // 系统号为56(微信团队官方帐号），
  // 29（未知，招行信用卡为29）
  // Flag是VerifyFlag解析后的值
  Flag int `json:"flag,omitempty"`

  // UserName每次登录都不一样，
  // 群聊帐号以@@开头，其他以@开头，内置帐号就直接是名字，如：
  // weixin（微信团队）/filehelper（文件传输助手）/fmessage(朋友消息推荐)
  UserName string `json:"-"`

  Bot *Bot `json:"-"`
}

func mapToContact(data map[string]interface{}, bot *Bot) *Contact {
  if data == nil || len(data) == 0 {
    return nil
  }
  ret := &Contact{Raw: data}
  ret.Bot = bot
  ret.Uin = bot.req.uin
  ret.Nickname = conv.String(data, "NickName")
  ret.UserName = conv.String(data, "UserName")
  if bot != nil && bot.Contacts != nil {
    if c := bot.Contacts.FindByUserName(ret.UserName); c != nil {
      ret.ID = c.ID
      ret.CreateTime = c.CreateTime
    }
  }
  if v, ok := data["VerifyFlag"]; ok {
    switch int(v.(float64)) {
    case 0:
      if (ret.UserName)[0:2] == "@@" {
        ret.Flag = ContactGroup
      } else if (ret.UserName)[0:1] == "@" {
        ret.Flag = ContactFriend
      } else {
        ret.Flag = ContactSystem
      }
    case 8, 24:
      ret.Flag = ContactMPS
    case 56:
      ret.Flag = ContactSystem
    default:
      ret.Flag = ContactUnknown
    }
  }
  return ret
}

func mapsToContacts(data []map[string]interface{}, bot *Bot) []*Contact {
  if data == nil || len(data) == 0 {
    return nil
  }
  ret := make([]*Contact, 0, len(data))
  for _, v := range data {
    contact := mapToContact(v, bot)
    if contact != nil {
      ret = append(ret, contact)
    }
  }
  return ret
}

func FindContactByID(id string) *Contact {
  var ret *Contact
  EachBot(func(b *Bot) bool {
    if b.Contacts != nil {
      if c := b.Contacts.FindByID(id); c != nil {
        ret = c
        return false
      }
    }
    return true
  })
  return ret
}

func (c *Contact) SendText(content string) error {
  if content == "" {
    return errInvalidArgs
  }
  return c.Bot.sendText(c.UserName, content)
}

func (c *Contact) SendImage(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", errInvalidArgs
  }
  return c.Bot.sendMedia(c.UserName, data, filename, MsgImage, sendImageURL)
}

func (c *Contact) SendVideo(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", errInvalidArgs
  }
  return c.Bot.sendMedia(c.UserName, data, filename, MsgVideo, sendVideoURL)
}

func (c *Contact) GetAttrString(attr string) string {
  return conv.String(c.Raw, attr)
}

func (c *Contact) GetAttrInt(attr string) int {
  return conv.Int(c.Raw, attr)
}

func (c *Contact) GetAttrUint64(attr string) uint64 {
  return conv.Uint64(c.Raw, attr)
}

func (c *Contact) GetAttrBool(attr string) bool {
  return conv.Bool(c.Raw, attr)
}

func (c *Contact) GetAttrBytes(attr string) []byte {
  if v, ok := c.Raw[attr]; ok {
    switch ret := v.(type) {
    case []byte:
      return ret
    case string:
      return []byte(ret)
    }
  }
  return nil
}

func parseRemarkToID(remark string) uint64 {
  ret, e := strconv.ParseUint(remark, 10, 64)
  if e == nil && ret > idInitial+idGeneralOffset {
    return ret
  }
  return 0
}
