package wechatbot

import (
  "strconv"
  "time"

  "github.com/kwf2030/commons/conv"
)

const (
  // 自带表情是文本消息，Content字段内容为：[奸笑]，
  // emoji表情也是文本消息，Content字段内容为：<span class="emoji emoji1f633"></span>，
  // 如果连同文字和表情一起发送，Content字段内容是文字和表情直接是混在一起，
  // 位置坐标也是文本消息，Content字段内容为：
  // 雨花台区雨花西路(德安花园东):/cgi-bin/mmwebwx-bin/webwxgetpubliclinkimg?url=xxx&msgid=741398718084560243&pictype=location
  MsgText = 1

  // 图片/照片消息，
  // Content字段内容为XML，Content字段内容为：
  // <?xml version="1.0"?>
  // <msg>
  // <img aeskey="" encryver="" cdnthumbaeskey="" cdnthumburl="" cdnthumblength=""
  //   cdnthumbheight="" cdnthumbwidth="" cdnmidheight="" cdnmidwidth="" cdnhdheight=""
  //   cdnhdwidth="" cdnmidimgurl="" length="" md5="" /><br/>
  // </msg>
  MsgImage = 3

  MsgVoice = 34

  // 被添加好友待验证，Content内容为：
  // <msg fromusername="kwf2030" encryptusername="v1_400be59c1cd145d71bcd4a389b68456833bfdba992d524a563494beed6def517@stranger"
  // fromnickname="kwf2030" content="我是客户"  shortpy="KWF2030" imagestatus="3" scene="30"
  // country="CN" province="Jiangsu" city="Nanjing" sign="" percard="1" sex="1" alias="" weibo=""
  // weibonickname="" albumflag="0" albumstyle="0" albumbgimgid="" snsflag="17" snsbgimgid=""
  // snsbgobjectid="0" mhash="21b5f7503b9728a74f69ffe2ac4a81b8"
  // mfullhash="21b5f7503b9728a74f69ffe2ac4a81b8"
  // bigheadimgurl="http://wx.qlogo.cn/mmhead/ver_1/as2mDdcIHonnibUkbSzmyAZ4eRPFv67M7IOLXhE4ULXQaRESLaNnLlsjHGvFuNXicnqYmxCXCZFjziaGQetfFyRhQ/0"
  // smallheadimgurl="http://wx.qlogo.cn/mmhead/ver_1/as2mDdcIHonnibUkbSzmyAZ4eRPFv67M7IOLXhE4ULXQaRESLaNnLlsjHGvFuNXicnqYmxCXCZFjziaGQetfFyRhQ/96"
  // ticket="v2_1604664f28c4e339b63f5299ef578d15350d9b02ee5b8137b0c568f5423fa5adfe843d9a7478dbf21395f26ae4567896f52e6cdd9f2971b81f06332c1f2c91bf@stranger"
  // opcode="2" googlecontact="" qrticket="" chatroomusername="" sourceusername="" sourcenickname="">
  // <brandlist count="0" ver="683212005"></brandlist>
  // </msg>
  MsgVerify = 37

  MsgFriendRecommend = 40

  // 名片消息，Content字段内容为：
  // <?xml version="1.0"?>
  // <msg bigheadimgurl="http://xxx" smallheadimgurl="http://xxx" username="v1_xxx@stranger" nickname=""
  // shortpy="" alias="" imagestatus="" scene="" province="" city="" sign="" sex="" certflag=""
  // certinfo="" brandIconUrl="" brandHomeUrl="" brandSubscriptConfigUrl="" brandFlags=""
  // regionCode="" antispamticket="v2_xxx@stranger" />
  MsgCard = 42

  // 拍摄（视频消息）
  MsgVideo = 43

  // 动画表情，
  // 包括官方表情包中的表情（Content字段无内容）和自定义的图片表情（Content字段内容为XML）
  MsgAnimEmotion = 47

  MsgLocation = 48

  // 公众号推送的链接，
  // 发送的文件也是链接消息，
  // 分享的链接（AppMsgType=1/3/5），
  // 红包（AppMsgType=2001）,
  // 收藏也是连接消息，
  // 实时位置共享也是链接消息，Content字段内容为：
  // <msg>
  // <appmsg appid="" sdkver="0">
  // <type>17</type>
  // <title><![CDATA[我发起了位置共享]]></title>
  // </appmsg>
  // <fromusername>kwf2030</fromusername>
  // </msg>
  MsgLink = 49

  MsgVoip = 50

  // 登录之后系统发送的初始化消息
  MsgInit = 51

  MsgVoipNotify = 52
  MsgVoipInvite = 53
  MsgVideoCall  = 62
  MsgNotice     = 9999

  // 系统消息，
  // 例如通过好友验证，系统会发送"你已添加了..."和"如果陌生人..."的消息，
  // 例如"实时位置共享已结束"的消息
  MsgSystem = 10000

  // 撤回消息，Content字段内容为：
  // <sysmsg type="revokemsg">
  // <revokemsg>
  // <session>Nickname</session>
  // <oldmsgid>1057920614</oldmsgid>
  // <msgid>2360839023010332147</msgid>
  // <replacemsg><![CDATA["Nickname" 撤回了一条消息]]></replacemsg>
  // </revokemsg>
  // </sysmsg>
  MsgRevoke = 10002
)

type Message struct {
  ID           string                 `json:"id,omitempty"`
  FromUserName string                 `json:"from_user_name,omitempty"`
  ToUserName   string                 `json:"to_user_name,omitempty"`
  FromUserID   string                 `json:"from_user_id,omitempty"`
  ToUserID     string                 `json:"to_user_id,omitempty"`
  Type         int                    `json:"type,omitempty"`
  URL          string                 `json:"url,omitempty"`
  Content      string                 `json:"content,omitempty"`
  CreateTime   time.Time              `json:"create_time,omitempty"`
  Raw          map[string]interface{} `json:"raw,omitempty"`
  Bot          *Bot                   `json:"-"`
}

func mapToMessage(data map[string]interface{}, bot *Bot) *Message {
  if data == nil || len(data) == 0 {
    return nil
  }
  ret := &Message{Raw: data}
  ret.Bot = bot
  ret.Type = conv.Int(data, "MsgType")
  ret.URL = conv.String(data, "Url")
  ret.Content = conv.String(data, "Content")
  if v, ok := data["CreateTime"]; ok {
    if x, ok := v.(float64); ok {
      ret.CreateTime = time.Unix(int64(x), 0)
    }
  }
  ret.ID = conv.String(data, "MsgId")
  if ret.ID == "" {
    if v, ok := data["NewMsgId"]; ok {
      if x, ok := v.(float64); ok {
        ret.ID = strconv.FormatUint(uint64(x), 10)
      }
    }
  }
  ret.FromUserName = conv.String(data, "FromUserName")
  if ret.FromUserName == "" && ret.Type == MsgVerify {
    ret.FromUserName = conv.String(conv.Map(data, "RecommendInfo"), "UserName")
  }
  ret.ToUserName = conv.String(data, "ToUserName")
  if bot != nil && bot.Contacts != nil {
    if c := bot.Contacts.FindByUserName(ret.FromUserName); c != nil {
      ret.FromUserID = c.ID
    }
    if c := bot.Contacts.FindByUserName(ret.ToUserName); c != nil {
      ret.ToUserID = c.ID
    }
  }
  return ret
}

func (msg *Message) GetFromContact() *Contact {
  if msg.Bot == nil || msg.Bot.Contacts == nil {
    return nil
  }
  return msg.Bot.Contacts.FindByUserName(msg.FromUserName)
}

func (msg *Message) GetToContact() *Contact {
  if msg.Bot == nil || msg.Bot.Contacts == nil {
    return nil
  }
  return msg.Bot.Contacts.FindByUserName(msg.ToUserName)
}

func (msg *Message) ReplyText(content string) error {
  if content == "" {
    return nil
  }
  return msg.Bot.sendText(msg.FromUserName, content)
}

func (msg *Message) ReplyImage(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", errInvalidArgs
  }
  return msg.Bot.sendMedia(msg.FromUserName, data, filename, MsgImage, sendImageURL)
}

func (msg *Message) ReplyVideo(data []byte, filename string) (string, error) {
  if len(data) == 0 || filename == "" {
    return "", errInvalidArgs
  }
  return msg.Bot.sendMedia(msg.FromUserName, data, filename, MsgVideo, sendVideoURL)
}

func (msg *Message) GetAttrString(attr string) string {
  return conv.String(msg.Raw, attr)
}

func (msg *Message) GetAttrInt(attr string) int {
  return conv.Int(msg.Raw, attr)
}

func (msg *Message) GetAttrUint64(attr string) uint64 {
  return conv.Uint64(msg.Raw, attr)
}

func (msg *Message) GetAttrBool(attr string) bool {
  return conv.Bool(msg.Raw, attr)
}

func (msg *Message) GetAttrBytes(attr string) []byte {
  if v, ok := msg.Raw[attr]; ok {
    switch ret := v.(type) {
    case []byte:
      return ret
    case string:
      return []byte(ret)
    }
  }
  return nil
}
