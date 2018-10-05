package wechatbot

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "math/rand"
  "net/http"
  "net/url"
  "os"
  "path"
  "strconv"
  "strings"

  "github.com/kwf2030/commons/conv"
)

var (
  verifyURL   = "/webwxverifyuser"
  remarkURL   = "/webwxoplog"
  logoutURL   = "/webwxlogout"
  contactsURL = "/webwxbatchgetcontact"
)

func (r *req) DownloadQRCode(dst string) (string, error) {
  resp, e := http.Get(fmt.Sprintf("%s/%s", qrURL, r.uuid))
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", errReq
  }
  data, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  if dst == "" {
    dst = path.Join(os.TempDir(), fmt.Sprintf("%d.jpg", rand.Int()))
  }
  e = ioutil.WriteFile(dst, data, os.ModePerm)
  if e != nil {
    return "", e
  }
  return dst, nil
}

func (r *req) DownloadAvatar(dst string) (string, error) {
  resp, e := r.client.Get(r.avatarURL)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", errReq
  }
  data, e := ioutil.ReadAll(resp.Body)
  if e != nil {
    return "", e
  }
  if dst == "" {
    dst = path.Join(os.TempDir(), fmt.Sprintf("%d.jpg", r.uin))
  }
  e = ioutil.WriteFile(dst, data, os.ModePerm)
  if e != nil {
    return "", e
  }
  return dst, nil
}

func (r *req) Verify(toUserName, ticket string) (map[string]interface{}, error) {
  if toUserName == "" || ticket == "" {
    return nil, errInvalidArgs
  }
  addr, _ := url.Parse(r.baseURL + verifyURL)
  q := addr.Query()
  q.Set("r", timestampString13())
  q.Set("pass_ticket", r.passTicket)
  addr.RawQuery = q.Encode()
  m := r.payload
  m["skey"] = r.skey
  m["Opcode"] = 3
  m["SceneListCount"] = 1
  m["SceneList"] = []int{33}
  m["VerifyContent"] = ""
  m["VerifyUserListSize"] = 1
  m["VerifyUserList"] = []map[string]string{
    {
      "Value":            toUserName,
      "VerifyUserTicket": ticket,
    },
  }
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errReq
  }
  return conv.ReadJSONToMap(resp.Body)
}

func (r *req) Remark(toUserName, remark string) (map[string]interface{}, error) {
  if toUserName == "" || remark == "" {
    return nil, errInvalidArgs
  }
  addr, _ := url.Parse(r.baseURL + remarkURL)
  q := addr.Query()
  q.Set("pass_ticket", r.passTicket)
  addr.RawQuery = q.Encode()
  m := r.payload
  m["UserName"] = toUserName
  m["CmdId"] = 2
  m["RemarkName"] = remark
  buf, _ := json.Marshal(m)
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errReq
  }
  return conv.ReadJSONToMap(resp.Body)
}

func (r *req) GetContacts(userNames []string) (map[string]interface{}, error) {
  if userNames == nil || len(userNames) == 0 {
    return nil, errInvalidArgs
  }
  addr, _ := url.Parse(r.baseURL + contactsURL)
  q := addr.Query()
  q.Set("type", "ex")
  q.Set("r", timestampString13())
  addr.RawQuery = q.Encode()
  list := make([]map[string]string, 0, len(userNames))
  for _, v := range userNames {
    item := make(map[string]string)
    item["UserName"] = v
    item["EncryChatRoomId"] = ""
    list = append(list, item)
  }
  m := r.payload
  m["Count"] = len(userNames)
  m["List"] = list
  buf, _ := json.Marshal(m)
  // 请求必须加上Content-Type和Cookies
  req, _ := http.NewRequest("POST", addr.String(), bytes.NewReader(buf))
  req.Header.Set("Referer", r.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", contentType)
  resp, e := r.client.Do(req)
  if e != nil {
    return nil, e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return nil, errReq
  }
  return conv.ReadJSONToMap(resp.Body)
}

func (r *req) Logout() {
  addr, _ := url.Parse(r.baseURL + logoutURL)
  q := addr.Query()
  q.Set("redirect", "1")
  q.Set("type", "1")
  q.Set("skey", r.skey)
  addr.RawQuery = q.Encode()
  form := url.Values{}
  form.Set("sid", r.sid)
  form.Set("uin", strconv.Itoa(r.uin))
  req, _ := http.NewRequest("POST", addr.String(), strings.NewReader(form.Encode()))
  req.Header.Set("Referer", r.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  resp, e := r.client.Do(req)
  if e != nil {
    return
  }
  resp.Body.Close()
}
