package oo

//common tool

import (
	"bytes"
	"io/ioutil"

	// "encoding/base64"
	"encoding/binary"
	"fmt"

	"math/big"
	"math/rand"
	"net"

	// "net/http"
	"os"
	"reflect"

	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"gopkg.in/go-playground/validator.v9"
)

const (
	DATE_FMT_STR      = "2006-01-02"
	DATE_TIME_FMT_STR = "2006-01-02 15:04:05"
)

func init() {
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)

	extra.RegisterTimeAsInt64Codec(time.Microsecond)

	extra.RegisterFuzzyDecoders()
}

func JoinArray(arr interface{}, sep string) string {
	v, ok := ToSlice(arr)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	s := ""
	for _, one := range v {
		prefix := sep
		if s == "" {
			prefix = ""
		}
		s = s + fmt.Sprintf("%s%v", prefix, one)
	}
	return s
}

func ToSlice(arr interface{}) ([]interface{}, bool) {
	v := reflect.ValueOf(arr)
	if v.Kind() != reflect.Slice {
		return nil, false
	}
	l := v.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = v.Index(i).Interface()
	}
	return ret, true
}
func StringSlice(ss interface{}) []string {
	if s, ok := ss.(string); ok {
		return []string{s}
	}
	return ss.([]string)
}
func Strings2Map(ss []string) (sm map[string]struct{}) {
	sm = map[string]struct{}{}
	for _, r := range ss {
		sm[r] = struct{}{}
	}
	return
}
func StringsUniq(ss []string, removes []string) []string {
	todels := Strings2Map(removes)

	n := 0
	for i, one := range ss {
		if _, ok := todels[one]; !ok {
			todels[one] = struct{}{}
			if n != i {
				ss[n] = ss[i]
			}
			n++
		}
	}
	if n != len(ss) {
		ss = ss[:n]
	}
	return ss
}

func InArray(needle interface{}, hystacks interface{}) bool {
	if harr, ok := ToSlice(hystacks); ok {
		for _, item := range harr {
			if item == needle {
				return true
			}
		}
	}
	return false
}

func ToStr(data interface{}) string {
	return fmt.Sprint(data)
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func JsonData(reqjson interface{}) []byte {
	if reqjson == nil {
		return []byte("")
	}

	data, err := jsoniter.Marshal(reqjson)
	if err != nil {
		return []byte("")
	}
	return data
}
func JsonDataIndent(reqjson interface{}) []byte {
	if reqjson == nil {
		return []byte("")
	}
	data, err := jsoniter.MarshalIndent(reqjson, "", "")
	if err != nil {
		return []byte("")
	}
	return data
}

func GetSvrmark(svrname string, serverid ...string) string {
	hostname, _ := os.Hostname()
	if pidx := strings.Index(string(hostname), "."); pidx > 0 {
		hostname = string([]byte(hostname)[:pidx-1])
	}
	if len(serverid) > 0 && len(serverid[0]) > 0 {
		return fmt.Sprintf("%s-%s", svrname, serverid[0])
	}
	pid := os.Getpid()
	return fmt.Sprintf("%s-%s-%d", hostname, svrname, pid)
}

// data source
func Gwei2Wei(gw int64) *big.Int {
	bgw := big.NewInt(gw)
	bw := big.NewInt(1000000000)

	return bw.Mul(bgw, bw)
}
func Wei2Gwei(bw *big.Int) int64 {
	bgw := big.NewInt(1000000000)
	return bgw.Div(bw, bgw).Int64()
}

func IP2Uint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func Uint32ToIP(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d", ip>>24, ip<<8>>24, ip<<16>>24, ip<<24>>24)
}

func Ts2Fmt(ts int64) string {
	return time.Unix(ts, 0).Format(DATE_TIME_FMT_STR)
}
func TimeNowUnix() int64 {
	return time.Now().Unix()
}

func Fmt2Ts(s string) int64 {
	v, err := time.ParseInLocation(DATE_TIME_FMT_STR, s, time.Local)
	if err != nil {
		return 0
	}
	return v.Unix()
}

func Fmt2Time(s string) (ret time.Time, err error) {
	ret, err = time.ParseInLocation(DATE_TIME_FMT_STR, s, time.Local)
	return
}

func Str2Time(str, format string) (ret time.Time, err error) {
	ret, err = time.ParseInLocation(format, str, time.Local)
	return
}

func Str2Int(str string) (i64 int64) {
	i64, _ = strconv.ParseInt(str, 10, 64)

	return
}

/*
type SliceHeader struct {
	Data uintptr
	Len  int
	Cap  int   //
  }
type StringHeader struct {
	Data uintptr
	Len  int
  }
*/
func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
	// return []byte(s)
}
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

type TagOptions struct {
	Skip      bool // "-"
	Name      string
	Omitempty bool
	Omitzero  bool
}

func GetTagOptions(tag reflect.StructTag, tagname string) TagOptions {
	t := tag.Get(tagname)
	if t == "-" {
		return TagOptions{Skip: true}
	}
	var opts TagOptions
	parts := strings.Split(t, ",")
	opts.Name = strings.Trim(parts[0], " ")
	for _, s := range parts[1:] {
		switch strings.Trim(s, " ") {
		case "omitempty":
			opts.Omitempty = true
		case "omitzero":
			opts.Omitzero = true
		}
	}
	return opts
}

func LowerCaseWithUnderscores(name string) string {
	newName := []rune{}
	for i, c := range name {
		if i == 0 {
			newName = append(newName, unicode.ToLower(c))
		} else {
			if unicode.IsUpper(c) {
				newName = append(newName, '_')
				newName = append(newName, unicode.ToLower(c))
			} else {
				newName = append(newName, c)
			}
		}
	}
	return string(newName)
}
func UpperCaseWithNoUnderscores(name string) string {
	newName := []rune{}
	under_flag := false
	for i, c := range name {
		if c == '_' {
			under_flag = true
			continue
		}
		if i == 0 || under_flag {
			c = unicode.ToUpper(c)
			under_flag = false
		}

		newName = append(newName, c)
	}
	return string(newName)
}

func Struct2Map(obj interface{}, args ...interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		sqltag := GetTagOptions(t.Field(i).Tag, "sqler")
		if sqltag.Skip || sqltag.Name == "skips" ||
			(len(args) > 0 && InArray(t.Field(i).Name, args[0])) {
			continue
		}
		field_name := t.Field(i).Name
		dbtag := GetTagOptions(t.Field(i).Tag, "db")
		if dbtag.Name != "" {
			field_name = dbtag.Name
		}
		data[field_name] = v.Field(i).Interface()
	}
	return data
}

func Struct2Fields(obj interface{}) (fields string) {
	t := reflect.TypeOf(obj)

	for i := 0; i < t.NumField(); i++ {
		field_name := t.Field(i).Name
		dbtag := GetTagOptions(t.Field(i).Tag, "db")
		if dbtag.Name != "" {
			field_name = dbtag.Name
		}

		if "" == fields {
			fields = field_name
		} else {
			fields += "," + field_name
		}
	}

	return
}

func randStr(size int, kind int) []byte {
	var fontKinds = [][]int{{10, 48}, {26, 97}, {26, 65}}
	ikind, result := kind, make([]byte, size)
	isAll := kind > 2 || kind < 0
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		if isAll {
			ikind = rand.Intn(3)
		}
		scope, base := fontKinds[ikind][0], fontKinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return result
}

func IntToBytes(num interface{}) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, num)
	return bytesBuffer.Bytes()
}

func BytesToInt(buf []byte) int64 {
	bytesBuffer := bytes.NewBuffer(buf)
	var num int64
	binary.Read(bytesBuffer, binary.BigEndian, &num)
	return num
}

func JsonMarshal(v interface{}) (data []byte, err error) {
	data, err = jsoniter.Marshal(v)
	return
}

func JsonUnmarshal(data []byte, v interface{}) (err error) {
	err = jsoniter.Unmarshal(data, v)
	return
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func LoadFile(fpath string) (data []byte, err error) {
	data, err = ioutil.ReadFile(fpath)
	return
}

func PtoT(capacity int64) float64 {
	return float64(capacity*10000/1024) / 10000
}

func SscanfEx(sep, str string, format string, a ...interface{}) (n int, err error) {
	if "" != sep {
		str = strings.Replace(str, sep, " ", -1)
		format = strings.Replace(format, sep, " ", -1)
	}
	return fmt.Sscanf(str, format, a...)
}

var validate = validator.New()

func ValidateStruct(val interface{}) (err error) {
	err = validate.Struct(val)
	if nil != err {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return
		}
		var str string
		for _, err2 := range err.(validator.ValidationErrors) {
			tmp := fmt.Sprintf("%s %s %s [%v]",
				err2.StructField(), err2.Tag(), err2.Param(), err2.Value())
			if "" != str {
				str += " | "
			}
			str += tmp
		}
		err = fmt.Errorf(str)
	}
	return
}

func JsonUnmarshalValidate(data []byte, val interface{}) (err error) {
	defer func(p *error) {
		if nil == *p {
			err = ValidateStruct(val)
		}
	}(&err)

	if 0 == len(data) {
		return
	}

	if nil == val {
		err = fmt.Errorf("the result pointer is nil")
	}

	err = jsoniter.Unmarshal(data, val)
	if nil != err {
		return
	}

	return
}
