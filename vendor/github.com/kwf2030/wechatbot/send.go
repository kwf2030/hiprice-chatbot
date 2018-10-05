package wechatbot

import (
  "bytes"
  "crypto/md5"
  "encoding/json"
  "fmt"
  "mime"
  "mime/multipart"
  "net/http"
  "net/url"
  "strconv"
  "strings"
  "time"

  "github.com/kwf2030/commons/conv"
  "github.com/kwf2030/commons/times"
)

const (
  sendTextURL    = "/webwxsendmsg"
  sendEmotionURL = "/webwxsendemoticon"
  sendImageURL   = "/webwxsendmsgimg"
  sendVideoURL   = "/webwxsendvideomsg"
  uploadURL      = "/webwxuploadmedia"
)

const dtFormat = "Mon Jan 02 2006 15:04:05 GMT-0700（中国标准时间）"

const chunk = 512 * 1024

func (r *req) SendText(toUserName, content string) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.baseURL + sendTextURL)
  q := addr.Query()
  q.Set("pass_ticket", r.passTicket)
  addr.RawQuery = q.Encode()
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + randStringN(4)
  params := map[string]interface{}{
    "Type":         MsgText,
    "Content":      content,
    "FromUserName": r.userName,
    "ToUserName":   toUserName,
    "LocalID":      s,
    "ClientMsgId":  s,
  }
  m := r.payload
  m["Scene"] = 0
  m["Msg"] = params
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

func (r *req) SendMedia(toUserName, mediaID string, msgType int, sendURL string) (map[string]interface{}, error) {
  addr, _ := url.Parse(r.baseURL + sendURL)
  q := addr.Query()
  q.Set("fun", "async")
  q.Set("f", "json")
  q.Set("pass_ticket", r.passTicket)
  addr.RawQuery = q.Encode()
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + randStringN(4)
  params := map[string]interface{}{
    "Type":         msgType,
    "MediaId":      mediaID,
    "FromUserName": r.userName,
    "ToUserName":   toUserName,
    "LocalID":      s,
    "ClientMsgId":  s,
    "Content":      "",
  }
  m := r.payload
  m["Scene"] = 0
  m["Msg"] = params
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

// data是上传的数据，如果大于chunk则按chunk分块上传，
// filename是文件名（非文件路径，用来检测文件类型和设置上传文件名，如1.png）
func (r *req) UploadMedia(toUserName string, data []byte, filename string) (string, error) {
  l := len(data)
  addr, _ := url.Parse(r.baseURL + uploadURL)
  addr.Host = "file." + addr.Host
  q := addr.Query()
  q.Set("f", "json")
  addr.RawQuery = q.Encode()

  mt := "application/octet-stream"
  i := strings.LastIndex(filename, ".")
  if i != -1 {
    t := mime.TypeByExtension(filename[i:])
    if t != "" {
      mt = t
    }
  }

  var mt2 string
  switch mt[:strings.Index(mt, "/")] {
  case "image":
    mt2 = "pic"
  case "video":
    mt2 = "video"
  default:
    mt2 = "doc"
  }

  hash := fmt.Sprintf("%x", md5.Sum(data))
  n, _ := strconv.ParseInt(timestampString13(), 10, 32)
  s := strconv.FormatInt(n<<4, 10) + randStringN(4)
  m := r.payload
  m["UploadType"] = 2
  m["ClientMediaId"] = s
  m["TotalLen"] = l
  m["DataLen"] = l
  m["StartPos"] = 0
  m["MediaType"] = 4
  m["FromUserName"] = r.userName
  m["ToUserName"] = toUserName
  m["FileMd5"] = hash
  req, _ := json.Marshal(m)

  info := &uploadInfo{
    addr:         addr.String(),
    filename:     filename,
    md5:          hash,
    mime:         mt,
    mediaType:    mt2,
    req:          string(req),
    fromUserName: r.userName,
    toUserName:   toUserName,
    dataTicket:   r.cookie("webwx_data_ticket"),
    totalLen:     l,
    wuFile:       r.wuFile,
    chunks:       0,
    chunk:        0,
    data:         nil,
  }
  defer func() { r.wuFile++ }()

  var mediaID string
  var err error
  if l <= chunk {
    info.data = data
    mediaID, err = r.uploadChunk(info)
  } else {
    m := l / chunk
    n := l % chunk
    if n == 0 {
      info.chunks = m
    } else {
      info.chunks = m + 1
    }
    for i := 0; i < m; i++ {
      s := i * chunk
      e := s + chunk
      info.chunk = i
      info.data = data[s:e]
      mediaID, err = r.uploadChunk(info)
      if err != nil {
        break
      }
    }
    if err == nil && n != 0 {
      info.chunk++
      info.data = data[l-n:]
      mediaID, err = r.uploadChunk(info)
    }
  }
  return mediaID, err
}

func (r *req) uploadChunk(info *uploadInfo) (string, error) {
  var buf bytes.Buffer
  w := multipart.NewWriter(&buf)
  w.WriteField("id", fmt.Sprintf("WU_FILE_%d", info.wuFile))
  w.WriteField("name", info.filename)
  w.WriteField("type", info.mime)
  w.WriteField("lastModifiedDate", times.Now().Add(time.Hour * -24).Format(dtFormat))
  w.WriteField("size", strconv.Itoa(info.totalLen))
  if info.chunks > 0 {
    w.WriteField("chunks", strconv.Itoa(info.chunks))
    w.WriteField("chunk", strconv.Itoa(info.chunk))
  }
  w.WriteField("mediatype", info.mediaType)
  w.WriteField("uploadmediarequest", info.req)
  w.WriteField("webwx_data_ticket", info.dataTicket)
  w.WriteField("pass_ticket", r.passTicket)
  fw, e := w.CreateFormFile("filename", info.filename)
  if e != nil {
    return "", e
  }
  if _, e = fw.Write(info.data); e != nil {
    return "", e
  }
  w.Close()

  req, _ := http.NewRequest("POST", info.addr, &buf)
  req.Header.Set("Referer", r.referer)
  req.Header.Set("User-Agent", userAgent)
  req.Header.Set("Content-Type", w.FormDataContentType())
  resp, e := r.client.Do(req)
  if e != nil {
    return "", e
  }
  defer resp.Body.Close()
  if resp.StatusCode != http.StatusOK {
    return "", errReq
  }
  ret, e := conv.ReadJSONToMap(resp.Body)
  if e != nil {
    return "", e
  }
  br := conv.Map(ret, "BaseResponse")
  rt := conv.Int(br, "Ret")
  if rt != 0 {
    return "", errReq
  }
  return conv.String(ret, "MediaId"), nil
}

type uploadInfo struct {
  addr         string
  filename     string
  md5          string
  mime         string
  mediaType    string
  req          string
  fromUserName string
  toUserName   string
  dataTicket   string
  totalLen     int
  wuFile       int
  chunks       int
  chunk        int
  data         []byte
}
