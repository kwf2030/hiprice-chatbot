package wechatbot

import (
  "strconv"

  "github.com/kwf2030/commons/times"
)

func timestamp() int64 {
  return times.Now().UnixNano()
}

func timestampStringN(l int, prefix, suffix string) string {
  s := strconv.FormatInt(timestamp(), 10)
  if len(s) <= l {
    return prefix + s + suffix
  }
  return prefix + s[:l] + suffix
}

func timestampString10() string {
  return timestampStringN(10, "", "")
}

func timestampString13() string {
  return timestampStringN(13, "", "")
}

func deviceID() string {
  return timestampStringN(15, "e", "")
}

func randStringN(l int) string {
  s := strconv.FormatInt(timestamp(), 10)
  if len(s) <= l {
    return s
  }
  return s[len(s)-l:]
}
