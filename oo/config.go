package oo

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"reflect"
	"strings"
	"sync"
)

type Config struct {
	sss  sync.Map //map[string]string
	cfgs sync.Map //map[string]interface{} //"xx.xx": []bool/int64/string
}

var GConfig *Config

func (c *Config) Int64(sesskey string, defs ...int64) int64 {
	if v, ok := c.cfgs.Load(sesskey); ok {
		ret, _ := v.(int64)
		return ret
	}
	defval := int64(0)
	if len(defs) > 0 {
		defval = defs[0]
	}

	return defval
}

func (c *Config) Bool(sesskey string, defs ...bool) bool {
	if v, ok := c.cfgs.Load(sesskey); ok {
		ret, _ := v.(bool)
		return ret
	}
	defval := bool(false)
	if len(defs) > 0 {
		defval = defs[0]
	}

	return defval
}

func (c *Config) String(sesskey string, defs ...string) string {
	if v, ok := c.cfgs.Load(sesskey); ok {
		ret, _ := v.(string)
		return ret
	}
	defval := string("")
	if len(defs) > 0 {
		defval = defs[0]
	}
	return defval
}

func (c *Config) Int64Array(sesskey string, defs ...[]int64) []int64 {
	if v, ok := c.cfgs.Load(sesskey); ok {
		var ret []int64
		if reflect.ValueOf(v).Kind() != reflect.Slice {
			if ret1, ok := v.(int64); ok {
				ret = []int64{ret1}
			}
		} else {
			for _, v1 := range v.([]interface{}) {
				if ret1, ok := v1.(int64); ok {
					ret = append(ret, ret1)
				}
			}
		}
		// fmt.Printf("Int64Array: %v, %v\n", v, ret)
		return ret
	}
	defval := []int64{}
	if len(defs) > 0 {
		defval = defs[0]
	}
	return defval
}

func (c *Config) BoolArray(sesskey string, defs ...[]bool) []bool {
	if v, ok := c.cfgs.Load(sesskey); ok {
		var ret []bool
		if reflect.ValueOf(v).Kind() != reflect.Slice {
			if ret1, ok := v.(bool); ok {
				ret = []bool{ret1}
			}
		} else {
			for _, v1 := range v.([]interface{}) {
				if ret1, ok := v1.(bool); ok {
					ret = append(ret, ret1)
				}
			}
		}
		// fmt.Printf("BoolArray: %v, %v\n", v, ret)
		return ret
	}
	defval := []bool{}
	if len(defs) > 0 {
		defval = defs[0]
	}
	return defval
}

func (c *Config) StringArray(sesskey string, defs ...[]string) []string {
	if v, ok := c.cfgs.Load(sesskey); ok {
		var ret []string
		if reflect.ValueOf(v).Kind() != reflect.Slice {
			if ret1, ok := v.(string); ok {
				ret = []string{ret1}
			}
		} else {
			for _, v1 := range v.([]interface{}) {
				if ret1, ok := v1.(string); ok {
					ret = append(ret, ret1)
				}
			}
		}
		// fmt.Printf("StringArray: %v, %v\n", v, ret)
		return ret
	}
	defval := []string{}
	if len(defs) > 0 {
		defval = defs[0]
	}
	return defval
}

func (c *Config) SessDecode(sess string, pdata interface{}) error {
	if v, ok := c.sss.Load(sess); ok {
		vs, _ := v.(string)
		_, err := toml.Decode(vs, pdata)
		return err
	}
	return NewError("no sess: %s", sess)
}

func (c *Config) SessDecodeMap(cfgmap map[string]interface{}) error {
	for sess, pdata := range cfgmap {
		if err := c.SessDecode(sess, pdata); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) SetSess(sess string, v interface{}) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		v = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}
	tv := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	for i := 0; i < tv.NumField(); i++ {
		key := tv.Field(i).Name
		t := GetTagOptions(tv.Field(i).Tag, "toml")
		if t.Name != "" {
			key = t.Name
		}
		sesskey := fmt.Sprintf("%s.%s", sess, key)
		c.cfgs.Store(sesskey, rv.Field(i).Interface())
	}

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(v); err == nil {
		c.sss.Store(sess, buf.String())
	}
}

func (c *Config) SetValue(sesskey string, v interface{}) {
	c.cfgs.Store(sesskey, v)
}

func (c *Config) SaveToml(fname string) error {
	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	//trans to map[string]map[string]interface{}
	tmp := map[string]map[string]interface{}{}
	c.cfgs.Range(func(sesskey1, v interface{}) bool {
		sesskey, _ := sesskey1.(string)
		sk := strings.Split(sesskey, ".")
		if len(sk) != 2 {
			return true
		}
		if _, ok := tmp[sk[0]]; !ok {
			tmp[sk[0]] = make(map[string]interface{})
		}
		tmp[sk[0]][sk[1]] = v
		return true
	})

	return toml.NewEncoder(f).Encode(tmp)
}

func convet_to_cfg(c *Config, tmp map[string]map[string]interface{}) {
	buf := new(bytes.Buffer)
	for sess, ss := range tmp {
		buf.Reset()
		if err := toml.NewEncoder(buf).Encode(ss); err == nil {
			c.sss.Store(sess, buf.String())
		}

		for key, val := range ss {
			sesskey := fmt.Sprintf("%s.%s", sess, key)
			c.cfgs.Store(sesskey, val)
		}
	}
}
func InitConfig(cfg_files interface{}, fn func(c *Config)) (*Config, error) {
	c := &Config{
		// sss:  make(map[string]string),
		// cfgs: make(map[string]interface{}),
	}
	var cfgs []string
	if cfg, ok := cfg_files.(string); ok {
		cfgs = []string{cfg}
	} else if cfgs1, ok := cfg_files.([]string); ok {
		for _, cfg := range cfgs1 {
			cfgs = append(cfgs, cfg)
		}
	}

	f := func(c *Config) {
		c.sss = sync.Map{}
		c.cfgs = sync.Map{}

		for _, f := range cfgs {
			tmp := map[string]map[string]interface{}{}
			_, err := toml.DecodeFile(f, &tmp)
			if err != nil {
				LogD("Not found config: %s, err=%v", f, err)
				continue
			}
			convet_to_cfg(c, tmp)
		}
	}

	f(c)

	if fn != nil {
		sig := NewSignalHandler(SigHup)
		go func() {
			for {
				select {
				case <-sig.GetChan():
					LogD("Get a SigHup.")
					f(c)

					fn(c)
				}
			}
		}()
	}

	return c, nil
}
