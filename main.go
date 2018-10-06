package main

import (
  "database/sql"
  "encoding/json"
  "fmt"
  "os"
  "os/signal"
  "strings"
  "time"

  "github.com/go-sql-driver/mysql"
  "github.com/kwf2030/commons/beanstalk"
  "github.com/kwf2030/commons/boltdb"
  "github.com/kwf2030/commons/times"
  "github.com/rs/zerolog"
  "errors"
)

const Version = "1.0.1"

var (
  bucketUserID  = []byte("user_id")
  bucketMsgSend = []byte("msg_send")
  bucketVar     = []byte("var")

  loopChan = make(chan struct{})

  logFile *os.File
  logger  *zerolog.Logger

  db *sql.DB
  kv *boltdb.KVStore
)

func main() {
  file := "conf.yaml"
  if len(os.Args) == 2 {
    file = os.Args[1]
  }
  e := LoadConf(file)
  if e != nil {
    panic(e)
  }

  initLogger()
  defer logFile.Close()
  logger.Info().Msg("Hiprice ChatBot " + Version)

  initDB()
  defer db.Close()

  initKV()
  defer kv.Close()

  go launchServer()
  go redirectHTTP()

  go run()
  loopChan <- struct{}{}

  s := make(chan os.Signal, 1)
  signal.Notify(s, os.Interrupt)
  <-s
}

func initLogger() {
  dir := Conf.Log.Dir
  e := os.MkdirAll(dir+"/dump", os.ModePerm)
  if e != nil {
    panic(e)
  }
  l := zerolog.DebugLevel
  switch strings.ToLower(Conf.Log.Level) {
  case "info":
    l = zerolog.InfoLevel
  case "warn":
    l = zerolog.WarnLevel
  case "error":
    l = zerolog.ErrorLevel
  case "fatal":
    l = zerolog.FatalLevel
  case "disable":
    l = zerolog.Disabled
  }
  zerolog.SetGlobalLevel(l)
  zerolog.TimeFieldFormat = ""
  if logFile != nil {
    logFile.Close()
  }
  logFile, _ = os.Create(fmt.Sprintf("%s/chatbot_%s.log", dir, times.NowStrFormat(times.DateFormat3)))
  lg := zerolog.New(logFile).Level(l).With().Timestamp().Logger()
  logger = &lg
  now := times.Now()
  next := now.Add(time.Hour * 24)
  next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
  time.AfterFunc(next.Sub(now), func() {
    logger.Info().Msg("create log file")
    go initLogger()
  })
}

func initDB() {
  for i := 0; i < 3; i++ {
    c := mysql.NewConfig()
    c.Net = "tcp"
    c.Addr = fmt.Sprintf("%s:%d", Conf.Database.Host, Conf.Database.Port)
    c.Collation = "utf8mb4_unicode_ci"
    c.User = Conf.Database.User
    c.Passwd = Conf.Database.Password
    c.DBName = Conf.Database.DB
    c.Loc = times.TimeZoneSH
    c.ParseTime = true
    c.Params = Conf.Database.Params
    var e error
    db, e = sql.Open("mysql", c.FormatDSN())
    if e != nil {
      logger.Info().Msg("database connect failed, will retry 30 seconds later")
      time.Sleep(time.Second * 30)
      continue
    }
    e = db.Ping()
    if e != nil {
      logger.Error().Err(e).Msg("database ping failed, will retry 10 seconds later")
      time.Sleep(time.Second * 10)
      continue
    }
    break
  }
  if db == nil {
    panic(errors.New("no database connection"))
  }
}

func initKV() {
  var e error
  kv, e = boltdb.Open("chatbot.db", "user_id", "msg_send", "var")
  if e != nil {
    panic(e)
  }
}

func run() {
  // 外层循环是定时任务
  for range loopChan {
    if isDayTime() {
      pushLocal()
    }
    conn, e := beanstalk.Dial(Conf.Beanstalk.Host, Conf.Beanstalk.Port)
    if e != nil {
      logger.Error().Err(e).Msg("ERR: Dial")
      scheduleNextTime()
      continue
    }
    // 内层循环是一直取任务直到没有为止
    for {
      id, data := reserveJob(conn)
      if id == "" {
        break
      }
      left := make(map[string]interface{}, 4)
      if bu, ok := data["by_user"]; ok {
        if bum, ok := bu.(map[string]interface{}); ok && len(bum) > 0 {
          m := pushByUser(bum)
          if len(m) > 0 {
            left["by_user"] = m
          }
        }
      }
      if bt, ok := data["by_text"]; ok {
        if btm, ok := bt.(map[string]interface{}); ok && len(btm) > 0 {
          m := pushByText(btm)
          if len(m) > 0 {
            left["by_text"] = m
          }
        }
      }
      if len(left) > 0 {
        left["create_time"] = data["create_time"]
        bytes, e := json.Marshal(left)
        if e != nil {
          logger.Error().Err(e).Msg("ERR: Marshal")
        }
        kv.UpdateV(bucketMsgSend, []byte(times.NowStr()), bytes)
      }
      e = conn.Delete(id)
      if e != nil {
        logger.Error().Err(e).Msg("ERR: Delete")
      }
    }
    e = conn.Quit()
    if e != nil {
      logger.Error().Err(e).Msg("ERR: Quit")
    }
    scheduleNextTime()
  }
}

func scheduleNextTime() {
  logger.Info().Msg("schedule next time")
  time.AfterFunc(time.Minute*time.Duration(Conf.Task.PollingInterval), func() {
    loopChan <- struct{}{}
  })
}

func isDayTime() bool {
  h := times.Now().Hour()
  return h > 7 && h < 23
}
