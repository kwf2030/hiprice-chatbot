package conv

import (
  "encoding/json"
  "io"
  "io/ioutil"
)

func MapToJSON(data map[string]interface{}) ([]byte, error) {
  ret, e := json.Marshal(data)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func JSONToMap(data []byte) (map[string]interface{}, error) {
  ret := make(map[string]interface{})
  e := json.Unmarshal(data, &ret)
  if e != nil {
    return nil, e
  }
  return ret, nil
}

func ReadJSON(r io.Reader, in interface{}) error {
  content, e := ioutil.ReadAll(r)
  if e != nil {
    return e
  }
  if content == nil || len(content) == 0 {
    return nil
  }
  return json.Unmarshal(content, in)
}

func ReadJSONToMap(r io.Reader) (map[string]interface{}, error) {
  content, e := ioutil.ReadAll(r)
  if e != nil {
    return nil, e
  }
  if content == nil || len(content) == 0 {
    return nil, nil
  }
  return JSONToMap(content)
}
