package main

import (
  "net/http"
  "time"

  "github.com/kwf2030/commons/conv"
)

const (
  StateWatch = iota
  StateUnWatch
)

const (
  NoScript   = -1
  NoValue    = -2
  RangePrice = -3
)

var watchListHandler = func(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodGet {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }
  u := r.URL.Query().Get("u")
  if u == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var uid string
  kv.QueryV(bucketUserID, []byte(u), func(k, v []byte, n int) error {
    uid = string(v)
    return nil
  })
  if uid == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  tx, _ := db.Begin()
  defer tx.Rollback()
  arr := make([]*Product, 0, 50)
  rows, e := tx.Query(`SELECT product_id, price, price_low, price_high FROM product_watch WHERE user_id=? AND state=?`, uid, StateWatch)
  for rows.Next() {
    p := NewProduct()
    e = rows.Scan(&p.ID, &p.WatchPrice, &p.WatchPriceLow, &p.WatchPriceHigh)
    if e != nil {
      continue
    }
    arr = append(arr, p)
  }
  rows.Close()
  if len(arr) == 0 {
    sendResp(w, 0, "", nil)
    return
  }
  for i := range arr {
    p := arr[i]
    r := tx.QueryRow(`SELECT _id, id, source, url, short_url, title, currency, price, price_low, price_high, stock, sales, category, update_time FROM product WHERE id=? LIMIT 1`, p.ID)
    r.Scan(&p.AID, &p.ID, &p.Source, &p.URL, &p.ShortURL,
      &p.Title, &p.Currency, &p.Price, &p.PriceLow, &p.PriceHigh,
      &p.Stock, &p.Sales, &p.Category, &p.UpdateTime)
  }
  sendResp(w, 0, "", map[string]interface{}{"products": arr})
}

var unwatchHandler = func(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }
  u := r.URL.Query().Get("u")
  if u == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var uid string
  kv.QueryV(bucketUserID, []byte(u), func(k, v []byte, n int) error {
    uid = string(v)
    return nil
  })
  if uid == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  m, _ := conv.ReadJSONToMap(r.Body)
  pid, ok := m["product_id"]
  if !ok || pid.(string) == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  db.Exec(`UPDATE product_watch SET state=? WHERE user_id=? AND product_id=?`, StateUnWatch, uid, pid)
  sendResp(w, 0, "", nil)
}

var remindHandler = func(w http.ResponseWriter, r *http.Request) {
  u := r.URL.Query().Get("u")
  if u == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var uid string
  kv.QueryV(bucketUserID, []byte(u), func(k, v []byte, n int) error {
    uid = string(v)
    return nil
  })
  if uid == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  switch r.Method {
  case "GET":
    pid := r.URL.Query().Get("p")
    if pid == "" {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    var rdo, rio int
    var rdv, riv float64
    r := db.QueryRow(`SELECT remind_decrease_option, remind_decrease_value, remind_increase_option, remind_increase_value FROM product_watch WHERE user_id=? AND product_id=? LIMIT 1`, uid, pid)
    e := r.Scan(&rdo, &rdv, &rio, &riv)
    if e != nil {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    sendResp(w, 0, "", map[string]interface{}{
      "remind_decrease_option": rdo,
      "remind_decrease_value":  rdv,
      "remind_increase_option": rio,
      "remind_increase_value":  riv,
    })

  case "POST":
    m, _ := conv.ReadJSONToMap(r.Body)
    pid, ok := m["product_id"]
    if !ok || pid.(string) == "" {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    rdo := int(m["remind_decrease_option"].(float64))
    rdv := m["remind_decrease_value"]
    rio := int(m["remind_increase_option"].(float64))
    riv := m["remind_increase_value"]
    db.Exec(`UPDATE product_watch SET remind_decrease_option=?, remind_decrease_value=?, remind_increase_option=?, remind_increase_value=? WHERE user_id=? AND product_id=?`, rdo, rdv, rio, riv, uid, pid)
    sendResp(w, 0, "", nil)

  default:
    w.WriteHeader(http.StatusMethodNotAllowed)
  }
}

var settingsHandler = func(w http.ResponseWriter, r *http.Request) {
  u := r.URL.Query().Get("u")
  if u == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }
  var uid string
  kv.QueryV(bucketUserID, []byte(u), func(k, v []byte, n int) error {
    uid = string(v)
    return nil
  })
  if uid == "" {
    w.WriteHeader(http.StatusBadRequest)
    return
  }

  switch r.Method {
  case "GET":
    var x int
    db.QueryRow(`SELECT disturb FROM user WHERE id=?`, uid).Scan(&x)
    sendResp(w, 0, "", map[string]interface{}{"disturb": x})

  case "POST":
    m, _ := conv.ReadJSONToMap(r.Body)
    v, ok := m["disturb"]
    if !ok {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    x := int(v.(float64))
    if x != 0 && x != 1 {
      w.WriteHeader(http.StatusBadRequest)
      return
    }
    db.Exec(`UPDATE user SET disturb=? WHERE id=?`, x, uid)
    sendResp(w, 0, "", nil)

  default:
    w.WriteHeader(http.StatusMethodNotAllowed)
  }
}

type Product struct {
  AID        uint64    `json:"_id"`
  ID         string    `json:"id,omitempty"`
  URL        string    `json:"url,omitempty"`
  ShortURL   string    `json:"short_url,omitempty"`
  Source     int       `json:"source,omitempty"`
  Title      string    `json:"title,omitempty"`
  Currency   int       `json:"currency,omitempty"`
  Price      float64   `json:"price,omitempty"`
  PriceLow   float64   `json:"price_low,omitempty"`
  PriceHigh  float64   `json:"price_high,omitempty"`
  Stock      int       `json:"stock,omitempty"`
  Sales      int       `json:"sales,omitempty"`
  Category   string    `json:"category,omitempty"`
  Comments   Comments  `json:"comments,omitempty"`
  UpdateTime time.Time `json:"update_time,omitempty"`

  WatchPrice     float64 `json:"watch_price,omitempty"`
  WatchPriceLow  float64 `json:"watch_price_low,omitempty"`
  WatchPriceHigh float64 `json:"watch_price_high,omitempty"`
}

func NewProduct() *Product {
  return &Product{
    Price: NoScript,
    Stock: NoScript,
    Sales: NoScript,
    Comments: Comments{
      Total: NoScript,
    },
    WatchPrice: NoScript,
  }
}

type Comments struct {
  Total  int `json:"total,omitempty"`
  Star5  int `json:"star5,omitempty"`
  Star4  int `json:"star4,omitempty"`
  Star3  int `json:"star3,omitempty"`
  Star2  int `json:"star2,omitempty"`
  Star1  int `json:"star1,omitempty"`
  Image  int `json:"image,omitempty"`
  Append int `json:"append,omitempty"`
}
