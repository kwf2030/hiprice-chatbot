package wechatbot

import (
  "errors"
  "fmt"
  "math/rand"
  "net/http"
  "net/http/cookiejar"
  "net/url"
  "os"
  "path"
  "runtime"
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/flow"
  "github.com/kwf2030/commons/times"
  "github.com/rs/zerolog"
  "golang.org/x/net/publicsuffix"
)

const (
  BotUnknown = iota

  BotCreated

  // 等待扫码确认或正在与服务器交互，此时还不能收发消息
  BotStarted

  // 登录成功，可以正常收发消息
  BotRunning

  // 停止/下线（手动、被动或异常）
  BotStopped
)

const (
  // 未扫码
  QRReady = iota

  // 已扫码未确认
  QRScanned

  // 已确认
  QRConfirmed

  // 超时
  QRTimeout
)

const (
  // 图片消息存放目录
  AttrDirImage = "wechatbot.attr.dir_image"

  // 语音消息存放目录
  AttrDirVoice = "wechatbot.attr.dir_voice"

  // 视频消息存放目录
  AttrDirVideo = "wechatbot.attr.dir_video"

  // 文件消息存放目录
  AttrDirFile = "wechatbot.attr.dir_file"

  // 头像存放路径
  AttrPathAvatar = "wechatbot.attr.path_avatar"

  // 持久化ID方案，如果禁用则Contact.ID永远不会有值，默认禁用，
  // 联系人持久化使用备注实现，群持久化使用群名称实现（改名会同步，同名没影响），公众号不会持久化
  AttrPersistentIDEnabled = "wechatbot.attr.persistent_id_enabled"

  // 未登录成功时会随机生成key，保证bots中有记录且可查询这个Bot
  attrBotPlaceHolder = "wechatbot.attr.bot_place_holder"

  // 起始ID
  attrInitialID = "wechatbot.attr.initial_id"
)

const (
  contentType = "application/json; charset=UTF-8"
  userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36"
)

var (
  errContactNotFound = errors.New("contact not found")

  errInvalidArgs = errors.New("invalid arguments")

  errInvalidState = errors.New("invalid state")

  errReq = errors.New("request failed")

  errResp = errors.New("response invalid")

  errTimeout = errors.New("timeout")
)

var (
  // 默认日志级别
  lLevel = zerolog.InfoLevel
  // 默认日志目录
  lDir = "log"
  // 默认数据目录
  dDir = "data"

  logFile *os.File
  logger  *zerolog.Logger

  // 初始化工作
  once sync.Once

  // 所有Bot，int=>*Bot
  // 若处于Created或Started状态，key是随机生成的，
  // 若处于Running和Stopped状态，key是uin，
  // 调用Bot.Release()会删除Bot
  bots sync.Map
)

func SetLogLevel(level string) {
  switch strings.ToLower(level) {
  case "debug":
    lLevel = zerolog.DebugLevel
  case "info":
    lLevel = zerolog.InfoLevel
  case "warn":
    lLevel = zerolog.WarnLevel
  case "error":
    lLevel = zerolog.ErrorLevel
  case "fatal":
    lLevel = zerolog.FatalLevel
  case "disabled":
    lLevel = zerolog.Disabled
  }
}

func SetDirs(logDir, dataDir string) {
  if logDir != "" {
    lDir = logDir
  }
  if dataDir != "" {
    dDir = dataDir
  }
}

func EachBot(f func(b *Bot) bool) {
  bots.Range(func(_, v interface{}) bool {
    if vv, ok := v.(*Bot); ok {
      return f(vv)
    }
    return true
  })
}

func AllBots() []*Bot {
  ret := make([]*Bot, 0, 2)
  EachBot(func(b *Bot) bool {
    ret = append(ret, b)
    return true
  })
  return ret
}

func RunningBots() []*Bot {
  ret := make([]*Bot, 0, 2)
  EachBot(func(b *Bot) bool {
    if b.State == BotRunning {
      ret = append(ret, b)
    }
    return true
  })
  return ret
}

func CountBots() int {
  i := 0
  EachBot(func(b *Bot) bool {
    i++
    return true
  })
  return i
}

func FindBotByUUID(uuid string) *Bot {
  if uuid == "" {
    return nil
  }
  var ret *Bot
  EachBot(func(b *Bot) bool {
    if b.req == nil || b.req.uuid != uuid {
      return true
    }
    ret = b
    return false
  })
  return ret
}

func FindBotByUin(uin int) *Bot {
  var ret *Bot
  EachBot(func(b *Bot) bool {
    if b.req == nil || b.req.uin != uin {
      return true
    }
    ret = b
    return false
  })
  return ret
}

func Destroy() {
  logger.Info().Msg("destroy")
  for _, b := range RunningBots() {
    b.Stop()
    b.Release()
  }
  bots = sync.Map{}
  once = sync.Once{}
  logger = nil
  if logFile != nil {
    logFile.Close()
    logFile = nil
  }
}

func doInit() {
  initLogger()
  now := times.Now()
  next := now.Add(time.Hour * 24)
  next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
  time.AfterFunc(next.Sub(now), func() {
    EachBot(func(b *Bot) bool {
      if b.State == BotRunning {
        b.updatePaths()
      }
      return true
    })
    logger.Info().Msg("create log file")
    go doInit()
  })
}

func initLogger() {
  e := os.MkdirAll(lDir, os.ModePerm)
  if e != nil {
    panic(e)
  }
  zerolog.SetGlobalLevel(lLevel)
  zerolog.TimeFieldFormat = ""
  if logFile != nil {
    logFile.Close()
  }
  logFile, _ = os.Create(fmt.Sprintf("%s/wechatbot_%s.log", lDir, times.NowStrFormat(times.DateFormat3)))
  lg := zerolog.New(logFile).Level(lLevel).With().Timestamp().Logger()
  logger = &lg
}

type Bot struct {
  Attr map[string]interface{}

  Self     *Contact
  Contacts *Contacts

  StartTime    time.Time
  StartTimeStr string
  StopTime     time.Time
  StopTimeStr  string

  // 状态转换是且仅是：
  // Created=>Started=>Running=>Stopped
  // Created=>Started=>Stopped
  State int

  opChan  chan *Op
  opChanI chan *op

  req *req
}

func CreateBot(enablePersistentID bool) *Bot {
  once.Do(func() {
    e := os.MkdirAll(dDir, os.ModePerm)
    if e != nil {
      panic(e)
    }
    doInit()
  })
  ch := make(chan *op, runtime.NumCPU()+1)
  bot := &Bot{
    Attr:    make(map[string]interface{}),
    State:   BotCreated,
    opChan:  make(chan *Op, cap(ch)),
    opChanI: ch,
    req:     newReq(ch),
  }
  // 未获取到uin之前key是随机的，
  // 登录失败或成功之后会删除这个key
  k := rand.Int()
  bot.Attr[attrBotPlaceHolder] = k
  bot.Attr[AttrPersistentIDEnabled] = enablePersistentID
  bots.Store(k, bot)
  logger.Debug().Msgf("bot created, key(rand)=%d", k)
  return bot
}

// qrChan为接收二维码URL的channel，
// 返回的channel用来接收事件和消息通知，
// Start方法会一直阻塞到登录成功可以开始收发消息为止
func (bot *Bot) Start(qrChan chan<- string) (<-chan *Op, error) {
  if qrChan == nil {
    return nil, errInvalidArgs
  }
  bot.State = BotStarted

  // 监听事件和消息
  go bot.dispatch()
  bot.req.initFlow()
  _, e := bot.req.flow.Start(qrChan)

  // 不管登录成功还是失败，都要把临时的kv删除
  bots.Delete(bot.Attr[attrBotPlaceHolder])

  if e != nil {
    logger.Error().Err(e).Msgf("bot start failed, key(rand)=%d", bot.Attr[attrBotPlaceHolder])
    // 登录Bot出现了问题或一直没扫描超时了
    bot.State = BotStopped
    bot.Release()
    return nil, e
  }

  t := times.Now()
  bot.StartTime = t
  bot.StartTimeStr = t.Format(times.DateTimeSFormat)
  bot.updatePaths()
  bot.State = BotRunning
  bots.Store(bot.Self.Uin, bot)
  logger.Debug().Msgf("bot started, key(uin)=%d", bot.Self.Uin)
  return bot.opChan, nil
}

func (bot *Bot) GetScanState() int {
  return bot.req.scanState
}

func (bot *Bot) GetAttrString(attr string) string {
  return conv.String(bot.Attr, attr)
}

func (bot *Bot) GetAttrInt(attr string) int {
  return conv.Int(bot.Attr, attr)
}

func (bot *Bot) GetAttrUint64(attr string) uint64 {
  return conv.Uint64(bot.Attr, attr)
}

func (bot *Bot) GetAttrBool(attr string) bool {
  return conv.Bool(bot.Attr, attr)
}

func (bot *Bot) GetAttrBytes(attr string) []byte {
  if v, ok := bot.Attr[attr]; ok {
    switch ret := v.(type) {
    case []byte:
      return ret
    case string:
      return []byte(ret)
    }
  }
  return nil
}

func (bot *Bot) updatePaths() error {
  dir := path.Join(dDir, strconv.Itoa(bot.Self.Uin), times.NowStrFormat(times.DateFormat))
  e := os.MkdirAll(dir, os.ModePerm)
  if e != nil {
    return e
  }

  di := path.Join(dir, "image")
  dvo := path.Join(dir, "voice")
  dvi := path.Join(dir, "video")
  df := path.Join(dir, "file")

  os.MkdirAll(di, os.ModePerm)
  os.MkdirAll(dvo, os.ModePerm)
  os.MkdirAll(dvi, os.ModePerm)
  os.MkdirAll(df, os.ModePerm)

  bot.Attr[AttrPathAvatar] = path.Join(path.Dir(dir), "avatar.jpg")
  bot.Attr[AttrDirImage] = di
  bot.Attr[AttrDirVoice] = dvo
  bot.Attr[AttrDirVideo] = dvi
  bot.Attr[AttrDirFile] = df
  return nil
}

func (bot *Bot) dispatch() {
  for o := range bot.opChanI {
    evt := &Op{What: o.What}
    switch o.What {
    case MsgOp:
      evt.Msg = mapToMessage(o.Data.(map[string]interface{}), bot)
      logger.Debug().Msgf("MsgOp, id=%s}", evt.Msg.ID)

    case ContactModOp:
      evt.Contact = bot.opContact(o.Data.(map[string]interface{}))
      logger.Debug().Msgf("ContactModOp, nickname=%s", evt.Contact.Nickname)

    case ContactDelOp:
      c := mapToContact(o.Data.(map[string]interface{}), bot)
      logger.Debug().Msgf("ContactDelOp, nickname=%s", evt.Contact.Nickname)
      if c.ID != "" {
        evt.Contact = c
        bot.Contacts.Remove(c.UserName)
      }

    case ContactSelfOp:
      bot.Self = mapToContact(o.Data.(map[string]interface{}), bot)
      logger.Debug().Msgf("ContactSelfOp, username=%s", bot.Self.UserName)

    case ContactListOp:
      bot.Contacts = initContacts(mapsToContacts(o.Data.([]map[string]interface{}), bot), bot)
      logger.Debug().Msgf("ContactListOp, size=%d", bot.Contacts.Size())

    case TerminateOp:
      logger.Debug().Msg("TerminateOp")
      bot.Stop()
    }
    // 事件转发
    bot.opChan <- evt
  }

  // opChanI不用关闭，在Logout之后，syncCheck请求会收到非零的响应，
  // 由它负责关闭opChanI（谁发送谁关闭的原则），
  // 但opChan是在dispatch的时候发送给调用方的，所以应该在此处关闭，
  // 不能放在Stop方法里，因为如果在Stop里面关闭了opChan，那就没法发送TerminateOp事件了，
  // 而且会引起"send on closed channel"的panic
  close(bot.opChan)
}

func (bot *Bot) opContact(m map[string]interface{}) *Contact {
  c := mapToContact(m, bot)
  logger.Debug().Msgf("opContact, nickname=%s", c.Nickname)
  if !bot.Attr[AttrPersistentIDEnabled].(bool) {
    bot.Contacts.Add(c)
    return c
  }
  switch c.Flag {
  case ContactFriend:
    // 如果ID是空，说明是新联系人
    if c.ID == "" {
      // 关闭好友验证的情况下，被添加好友时会收到此类消息，
      // ContactModOp会先于MsgOp事件发出，所以收到MsgOp时，该联系人一定已存在
      c.ID = strconv.FormatUint(bot.Contacts.NextID(), 10)
      c.CreateTime = times.Now()
      bot.req.Remark(c.UserName, c.ID)
      logger.Debug().Msgf("new contact, id=%s", c.ID)
    }
  case ContactGroup:

  case ContactSystem:
    n, ok := internalIDs[c.UserName]
    if !ok {
      n = uint64(len(internalIDs) + 1)
      internalIDs[c.UserName] = n
    }
    c.ID = strconv.FormatUint(bot.Contacts.initialID()+n, 10)
    logger.Debug().Msgf("new contact, username=%s", c.UserName)
  }
  bot.Contacts.Add(c)
  return c
}

func (bot *Bot) Stop() {
  bot.State = BotStopped
  t := times.Now()
  bot.StopTime = t
  bot.StopTimeStr = t.Format(times.DateTimeSFormat)
  bot.req.Logout()
  logger.Debug().Msgf("bot stopped, key(uin)=%d", bot.Self.Uin)
}

func (bot *Bot) Release() {
  bots.Delete(bot.req.uin)
  bot.req.reset()
  bot.req.flow = nil
  bot.req.client = nil
  bot.req = nil
  bot.Attr = nil
  bot.Self = nil
  bot.Contacts = nil
  bot.opChan = nil
  bot.opChanI = nil
  bot.State = BotUnknown
  if bot.Self != nil {
    uin := bot.Self.Uin
    logger.Debug().Msgf("bot released, key(uin)=%d", uin)
  } else {
    logger.Debug().Msgf("bot released")
  }
}

type op struct {
  What int
  Data interface{}
}

type Op struct {
  What    int
  Msg     *Message
  Contact *Contact
}

type req struct {
  op     chan<- *op
  flow   *flow.Flow
  client *http.Client
  *session
}

func newReq(op chan<- *op) *req {
  jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
  s := &session{}
  s.reset()
  ret := &req{
    session: s,
    op:      op,
    flow:    flow.NewFlow(0),
    client: &http.Client{
      Jar:     jar,
      Timeout: time.Minute * 2,
    },
  }
  return ret
}

func (r *req) initFlow() {
  uuid := &UUIDReq{r}
  scanState := &ScanStateReq{r}
  login := &LoginReq{r}
  init := &InitReq{r}
  statusNotify := &StatusNotifyReq{r}
  contactList := &ContactListReq{r}
  syn := &SyncReq{r}
  r.flow.AddLast(uuid, "uuid")
  r.flow.AddLast(scanState, "scan_state")
  r.flow.AddLast(login, "login")
  r.flow.AddLast(init, "init")
  r.flow.AddLast(statusNotify, "status_notify")
  r.flow.AddLast(contactList, "contact_list")
  r.flow.AddLast(syn, "sync")
}

func (r *req) cookie(key string) string {
  if key == "" {
    return ""
  }
  addr, _ := url.Parse(r.baseURL)
  arr := r.client.Jar.Cookies(addr)
  for _, c := range arr {
    if c.Name == key {
      return c.Value
    }
  }
  return ""
}

type session struct {
  referer       string
  host          string
  syncCheckHost string
  baseURL       string

  uuid        string
  redirectURL string
  uin         int
  sid         string
  skey        string
  passTicket  string
  payload     map[string]interface{}

  scanState int

  userName  string
  avatarURL string
  syncKey   map[string]interface{}

  wuFile int
}

func (s *session) reset() {
  s.referer = "https://wx.qq.com/"
  s.host = "wx.qq.com"
  s.syncCheckHost = "webpush.weixin.qq.com"
  s.baseURL = "https://wx.qq.com/cgi-bin/mmwebwx-bin"
  s.uuid = ""
  s.redirectURL = ""
  s.uin = 0
  s.sid = ""
  s.skey = ""
  s.passTicket = ""
  s.payload = nil
  s.scanState = QRReady
  s.userName = ""
  s.avatarURL = ""
  s.syncKey = nil
  s.wuFile = 0
}
