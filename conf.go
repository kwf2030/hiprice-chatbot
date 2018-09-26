package main

import (
  "io/ioutil"

  "gopkg.in/yaml.v2"
)

var Conf = &struct {
  Email     string        `yaml:"email"`
  Server    ServerConf    `yaml:"server"`
  MMS       MMSConf       `yaml:"mms"`
  Log       LogConf       `yaml:"log"`
  Beanstalk BeanstalkConf `yaml:"beanstalk"`
  Database  DatabaseConf  `yaml:"database"`
  Task      TaskConf      `yaml:"task"`
}{}

type ServerConf struct {
  Host     string   `yaml:"host"`
  Port     int      `yaml:"port"`
  Cert     string   `yaml:"cert"`
  Key      string   `yaml:"key"`
  User     string   `yaml:"user"`
  Password string   `yaml:"password"`
  Web      string   `yaml:"web"`
  Cors     []string `yaml:"cors"`
}

type MMSConf struct {
  Enabled      int    `yaml:"enabled"`
  EndpointHost int    `yaml:"endpoint_host"`
  EndpointPath string `yaml:"endpoint_path"`
  Sign         string `yaml:"sign"`
  TemplateCode string `yaml:"template_code"`
  Token        string `yaml:"token"`
  Tel          string `yaml:"tel"`
}

type LogConf struct {
  Dir   string `yaml:"dir"`
  Level string `yaml:"level"`
}

type BeanstalkConf struct {
  Host           string `yaml:"host"`
  Port           int    `yaml:"port"`
  ReserveTube    string `yaml:"reserve_tube"`
  ReserveTimeout int    `yaml:"reserve_timeout"`
}

type DatabaseConf struct {
  Host     string            `yaml:"host"`
  Port     int               `yaml:"port"`
  DB       string            `yaml:"db"`
  User     string            `yaml:"user"`
  Password string            `yaml:"password"`
  Params   map[string]string `yaml:"params"`
}

type TaskConf struct {
  PollingInterval int `yaml:"polling_interval"`
  MaxSend         int `yaml:"max_send"`
  MaxSendDelay    int `yaml:"max_send_delay"`
}

func LoadConf(file string) error {
  data, e := ioutil.ReadFile(file)
  if e != nil {
    return e
  }
  e = yaml.Unmarshal(data, Conf)
  if e != nil {
    return e
  }
  Conf.Server.Cors = append(Conf.Server.Cors, Conf.Server.Web)
  return nil
}
