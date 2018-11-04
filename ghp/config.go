package main

import (
  "bytes"
  "io"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
  "sync"
  "gopkg.in/yaml.v2"
)

type HttpServerConfig struct {
  Address string
  Port    uint
}

type Config struct {
  Gopath   string
  BuildDir string `yaml:"build-dir"`
  DevMode  bool   `yaml:"dev-mode"`
  PubDir   string `yaml:"pub-dir"`

  HttpServer []*HttpServerConfig `yaml:"http-server"`

  DirList struct {
    Enabled  bool
    Template string
  } `yaml:"dir-list"`

  Servlet struct {
    Preload   bool
    HotReload bool `yaml:"hot-reload"`
  }

  // ---- internal ----
  goProcEnv []string
  goProcEnvOnce sync.Once
}


func (c *Config) load(r io.Reader) error {
  data, err := ioutil.ReadAll(r)
  if err != nil {
    return err
  }
  data = bytes.Replace(data, []byte("${ghpdir}"), []byte(ghpdir), -1)
  r = bytes.NewReader(data)

  d := yaml.NewDecoder(r)
  d.SetStrict(true)
  return d.Decode(c)
}


func (c *Config) writeYaml(w io.Writer) error {
  // d, err := yaml.Marshal(c)
  // if err != nil {
  //   return err
  // }
  // d = bytes.Replace(d, []byte(ghpdir + "/"), []byte("${ghpdir}/"), -1)
  // d = bytes.Replace(d, []byte(ghpdir + "\n"), []byte("${ghpdir}\n"), -1)
  // d = bytes.Replace(d, []byte(ghpdir + "\""), []byte("${ghpdir}\""), -1)
  // r := bytes.NewReader(d)
  // _, err = r.WriteTo(w)
  // return err
  return yaml.NewEncoder(w).Encode(c)
}


func (c *Config) getGoProcEnv() []string {
  c.goProcEnvOnce.Do(func() {
    var env []string
    // configure env
    // TODO: only do this once (can use sync.Once)
    gopathPrefix := "GOPATH="
    for _, ent := range os.Environ() {
      if strings.HasPrefix(ent, gopathPrefix) {
        // prepend our explicit gopath to GOPATH
        GOPATH := ent[len(gopathPrefix):]
        ent = gopathPrefix + config.Gopath
        if len(GOPATH) > 0 {
          ent += string(filepath.ListSeparator) + GOPATH
        }
      }
      env = append(env, ent)
    }
    c.goProcEnv = env
  })
  return c.goProcEnv
}



func openUserConfigFile() (*os.File, error) {
  // try different locations
  locations := []string{
    "ghp.yaml",
    "ghp.yml",
  }
  for _, name := range locations {
    f, err := os.Open(name)
    if err == nil {
      return f, nil
    }
    if !os.IsNotExist(err) {
      return nil, err
    }
  }
  return nil, nil
}


func loadConfig() (*Config, error) {
  c := &Config{}

  // load base config
  baseConfigName := pjoin(ghpdir, "misc", "ghp.yaml")
  f, err := os.Open(baseConfigName)
  if err != nil {
    if os.IsNotExist(err) {
      err = errorf("base config file not found: %s", baseConfigName)
    }
    return nil, err
  }
  defer f.Close()
  logf("loading config %q", f.Name())
  if err = c.load(f); err != nil {
    return nil, err
  }

  // load optional user config, which can override any config properties
  f, err = openUserConfigFile()
  if err != nil {
    return nil, err
  }
  if f != nil {
    defer f.Close()
    logf("loading config %q", f.Name())
    if err := c.load(f); err != nil {
      return nil, err
    }
  }

  // Canonicalize paths (preserves symlinks)
  c.Gopath = abspath(c.Gopath)
  c.PubDir = abspath(c.PubDir)
  c.BuildDir = abspath(c.BuildDir)
  c.DirList.Template = abspath(c.DirList.Template)

  return c, nil
}