package httputil

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "math/rand"
  "net/http"
  "time"
)

const shortURLEndpoint = "http://api.t.sina.com.cn/short_url/shorten.json?source=%s&url_long=%s"

var (
  shortURLApiKey    = []string{"3271760578", "1681459862"}
  shortURLApiKeyLen = int32(len(shortURLApiKey))

  rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func ShortenURL(addr string) string {
  resp, e := http.Get(fmt.Sprintf(shortURLEndpoint, shortURLApiKey[rnd.Int31n(shortURLApiKeyLen)], addr))
  if e != nil {
    return ""
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return ""
  }
  a := make([]map[string]interface{}, 1)
  content, _ := ioutil.ReadAll(resp.Body)
  json.Unmarshal(content, &a)
  if len(a) > 0 {
    if v, ok := a[0]["url_short"]; ok {
      return v.(string)
    }
  }
  return ""
}
