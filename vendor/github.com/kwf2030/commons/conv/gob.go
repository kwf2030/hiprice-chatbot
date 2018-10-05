package conv

import (
  "bytes"
  "encoding/gob"
  "io"
  "io/ioutil"
)

func MapToGob(data map[string]interface{}) ([]byte, error) {
  var buf bytes.Buffer
  e := gob.NewEncoder(&buf).Encode(data)
  if e != nil {
    return nil, e
  }
  return buf.Bytes(), nil
}

func GobToMap(data []byte) (map[string]interface{}, error) {
  ret := make(map[string]interface{})
  e := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func ReadGob(r io.Reader, in interface{}) error {
  content, e := ioutil.ReadAll(r)
  if e != nil {
    return e
  }
  if content == nil || len(content) == 0 {
    return nil
  }
  return gob.NewDecoder(bytes.NewBuffer(content)).Decode(in)
}

func ReadGobToMap(r io.Reader) (map[string]interface{}, error) {
  content, e := ioutil.ReadAll(r)
  if e != nil {
    return nil, e
  }
  if content == nil || len(content) == 0 {
    return nil, nil
  }
  return GobToMap(content)
}
