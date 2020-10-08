package oo

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/json-iterator/go"
	// "net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DataHandler = func(*WebSock, []byte) ([]byte, error)

type WebSock struct {
	Ws   *websocket.Conn
	Ch   *Channel
	Data interface{}

	recv_fn     DataHandler
	open_fn     func(*WebSock) error
	close_fn    func(*WebSock)
	interval_fn func(*WebSock, int64)

	co_sche    bool
	read_timeo int64

	//for rpc
	Sess          *Session
	ctxHandleMap  sync.Map
	defCtxHandler ReqCtxHandler

	create_ts   int64
	read_ts     int64
	write_ts    int64
	read_count  uint64
	write_count uint64
	read_bytes  uint64
	write_bytes uint64

	is_client bool
}

//chmgr should be == nil
func InitWebSock(w http.ResponseWriter, r *http.Request, chmgr *ChannelManager, recv_fn DataHandler) (c *WebSock, err error) {
	c = &WebSock{}

	var upgrader = websocket.Upgrader{
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	c.Ws, err = upgrader.Upgrade(w, r, http.Header{
		"Access-Control-Allow-Origin":      []string{"*"},
		"Access-Control-Allow-Credentials": []string{"true"},
	})
	if err != nil {
		return nil, err
	}

	c.Ch = chmgr.NewChannel()
	c.Ch.Data = c
	c.create_ts = time.Now().Unix()
	c.recv_fn = recv_fn

	go c.serializedSend()

	return c, nil
}

func InitWsClient(Scheme string, Host string, Path string, chmgr *ChannelManager, recv_fn DataHandler) (*WebSock, error) {
	c := &WebSock{}

	// c.ori_host = Host

	vurl := &url.URL{Scheme: Scheme, Host: Host, Path: Path}
	if pp := strings.IndexByte(vurl.Path, '?'); pp >= 0 {
		vurl.RawQuery = string(vurl.Path[pp+1:])
		vurl.Path = string(vurl.Path[:pp])
	}

	c.Ch = chmgr.NewChannel()
	c.Ch.Data = vurl
	c.create_ts = time.Now().Unix()
	c.recv_fn = recv_fn

	c.is_client = true

	go c.serializedSend()

	return c, nil
}

func (c *WebSock) StartDial(chk_interval int64, max_reconnect int64) {
	vurl, _ := c.Ch.Data.(*url.URL)
	c.Ch.Data = c

	defer func() {
		if errs := recover(); errs != nil {
			LogW("recover StartDial %v. err=%v", vurl, errs)
		}
	}()

	if max_reconnect == int64(-1) {
		max_reconnect = int64(8388608)
	}
	checkfn := func(c *WebSock) {
		if c.Ws == nil {
			// LogD("to dial %v, host: %s", vurl, c.ori_host)
			timeo := int64(10)
			if timeo < c.read_timeo {
				timeo = c.read_timeo
			}
			dialer := &websocket.Dialer{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
					ServerName:         vurl.Host,
				},
				HandshakeTimeout: time.Second * time.Duration(timeo),
			}
			wconn, _, err := dialer.Dial(vurl.String(), http.Header{"Host": []string{vurl.Host}})
			if err != nil {
				StatChg("connect fail "+vurl.String(), 1)
				if c.Ch.Chmgr.DebugFlag > 0 {
					LogD("Failed to Dial %v, host: %s, err=%v", vurl, vurl.Host, err)
				}
			} else {
				StatChg("connect succ "+vurl.String(), 1)
				c.Ws = wconn
				// LogD("Succeed Dial %v, host: %s", vurl, vurl.Host) // remove for cli

				go c.RecvRequest()
			}
			max_reconnect--
		}
	}

	if chk_interval == 0 {
		chk_interval = 60
	}

	for max_reconnect >= 0 {

		checkfn(c)

		select {
		case <-c.Ch.IsClosed():
			return
		default:
			time.Sleep(time.Duration(chk_interval) * time.Second)
			// case <-time.After(time.Second * time.Duration(chk_interval)):
		}

	}
}

func (c *WebSock) SetIntervalHandler(fn func(c *WebSock, ms int64)) {
	c.interval_fn = fn
}
func (c *WebSock) SetOpenHandler(handler func(c *WebSock) error) {
	c.open_fn = handler
}
func (c *WebSock) SetCloseHandler(handler func(c *WebSock)) {
	c.close_fn = handler
}

func (c *WebSock) PeerAddr() (s string) {
	if c != nil && c.Ws != nil {
		if c.Sess != nil {
			s = c.Sess.Ip
		} else {
			s = c.Ws.RemoteAddr().String()
		}
	}
	return
}
func (c *WebSock) ConnId() uint64 {
	return c.Ch.GetSeq()
}

func (c *WebSock) ConnInfo() (s string) {
	if c != nil && c.Ws != nil {
		s += fmt.Sprintf("c=[%d,%s]; r=(%s,%d,%d); w=(%s,%d,%d)<-->%s", c.Ch.GetSeq(),
			Ts2Fmt(c.create_ts), Ts2Fmt(c.read_ts), c.read_count, c.read_bytes,
			Ts2Fmt(c.write_ts), c.write_count, c.write_bytes, c.PeerAddr())
	}
	// if c != nil && c.url != nil {
	// 	s += fmt.Sprintf("(%v)", c.url)
	// }
	return
}
func (c *WebSock) Close() {
	c.Ch.Close()

	if c.Ws != nil {
		c.Ws.Close()
		c.Ws = nil
	}
}

func (c *WebSock) OpenCoroutineFlag() {
	c.co_sche = true
}

func (c *WebSock) SetReadTimeout(read_timeo int64) {
	c.read_timeo = read_timeo
}
func (c *WebSock) IsReady() bool {
	return c.Ws != nil
}
func (c *WebSock) IsClient() bool {
	return c.is_client
}

func (c *WebSock) RecvRequest() {
	defer func() {
		if errs := recover(); errs != nil {
			LogW("recover RecvRequest %s. err=%v", c.ConnInfo(), errs)
		}
		if c.close_fn != nil {
			c.close_fn(c)
		}
		// LogD("Exit RecvRequest. ws:%p, %s", c, c.ConnInfo())
		if c.Ws != nil {
			c.Ws.Close()
			c.Ws = nil
		}
	}()

	if c.open_fn != nil {
		if err := c.open_fn(c); err != nil {
			LogD("failed to open conn :%s, err=%v", c.ConnInfo(), err)
			return
		}
	}

	// For:
	for c.Ws != nil {
		select {
		case <-c.Ch.IsClosed():
			return //break For
		default:
		}
		if c.read_timeo > 0 {
			dline := time.Now().Add(time.Second * time.Duration(c.read_timeo))
			c.Ws.SetReadDeadline(dline)
		}
		messageType, message, err := c.Ws.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				// LogD("Close websocket %s. err=%v", c.ConnInfo(), err) // remove for cli
			}
			return //break For
		}
		// StatChg("WSREAL RECV", 1)
		// StatChg(PERFSTAT_INCOMING_TRAFFIC, uint64(len(message)))

		c.read_ts = time.Now().Unix()
		c.read_count++
		c.read_bytes += uint64(len(message))

		if !(messageType == websocket.TextMessage || messageType == websocket.BinaryMessage) || len(message) < 11 {
			//skip len("{\"cmd\":\"a\"}") == 11
			LogD("NULL MSG: type=%d len=%d [%v] [%s]", messageType, len(message), message, Bytes2Str(message))
			StatChg("NULL Message", 1)
			continue
		}
		// StatChg(PERFSTAT_REQUEST, 1)

		call_fn := func() {
			defer func() {
				if errs := recover(); errs != nil {
					LogW("recover onRecvAsData %s. err=%v", c.ConnInfo(), errs)
				}
			}()
			ret_d, err := c.recv_fn(c, message)
			if err != nil {
				LogD("Failed to process %s err: %v", c.ConnInfo(), err)
			}
			if err == nil && len(ret_d) > 0 {
				if err = c.SendData(ret_d); err != nil {
					LogW("Failed to send websocket %s err: %v", c.PeerAddr(), err)
				}
			}
		}
		if c.co_sche {
			go call_fn()
		} else {
			call_fn()
		}

	}
}

func (c *WebSock) serializedSend() {
	defer func() {
		if errs := recover(); errs != nil {
			LogW("recover serializedSend %s. err=%v", c.ConnInfo(), errs)
		}
	}()

	t := time.NewTimer(1 * time.Second)
	defer t.Stop()

For:
	for {
		t.Stop()
		t = time.NewTimer(1 * time.Second)

		select {
		case m, ok := <-c.Ch.RecvChan():
			if !ok {
				// StatChg("send notok", 1)
				break For
			}

			if c.Ws == nil { //active connection
				StatChg("active wait", 1)
				// <-time.After(time.Second * 1)
				// time.Sleep(time.Second * 1)
				// if err := c.Ch.PushMsg(m); err != nil {
				// 	LogW("Failed to send %s -> %s err:%v", m, c.PeerAddr(), err)
				// }
				continue
			}

			data, ok := m.([]byte)
			if !ok || len(data) < 8 {
				LogW("Failed to convert msg: %s, ok:%v. m=%v", string(data), ok, m)
				continue
			}

			// StatChg("WSREAL SENT", 1)
			if err := c.Ws.WriteMessage(websocket.TextMessage, data); err != nil {
				LogD("write to %s err: %v", c.ConnInfo(), err)
				// StatChg("WSREAL SENT", -1)
			}
			c.write_ts = time.Now().Unix()
			c.write_count++
			c.write_bytes += uint64(len(data))
			// case <-time.After(1 * time.Second): //Less accurate
		case <-t.C:
			if c.interval_fn != nil && c.Ws != nil {
				ms := time.Now().UnixNano() / 1e6
				go c.interval_fn(c, ms)
			}
		}
	}
}

func (c *WebSock) SendData(data []byte) (err error) {
	err = c.Ch.PushMsg(data)
	return
}

//////////////////////////////////////////// for rpc ////////////////////////////////////////////

var s_sndid uint64
var s_sndmap sync.Map

type WsHandler = func(*WebSock, *RpcMsg) (*RpcMsg, error)

func RecvRpcFn(c *WebSock, message []byte) (ret_data []byte, err error) {
	rpcmsg := &RpcMsg{}
	if err = jsoniter.Unmarshal(message, rpcmsg); err != nil {
		LogD("%s json parse error. len=%d. msg=%s", c.ConnInfo(), len(message), Bytes2Str(message))
		return
	}
	// StatChg(fmt.Sprintf("%s%s", PERFSTAT_REQUEST_PREFIX, rpcmsg.Cmd), 1)
	// LogD("Get cmd req: %s", rpcmsg.Cmd)

	if rpcmsg.Typ == RpcTypeResp {
		if len(rpcmsg.Chk) > 0 && strings.Count(rpcmsg.Chk, ".") > 0 {
			if ch, ok := s_sndmap.Load(rpcmsg.Chk); ok {
				// LogD("get reply ch. %s, chk:%s", rpcmsg.Cmd, rpcmsg.Chk)
				ch.(chan *RpcMsg) <- rpcmsg
				// return
			}
		}
		return //nobody wait
	}

	fn := c.defCtxHandler
	if v, ok := c.ctxHandleMap.Load(rpcmsg.Cmd); ok {
		// LogD("Get push hand fn, %s", rpcmsg.Cmd)
		fn, _ = v.(func(*ReqCtx, *RpcMsg) (*RpcMsg, error))
	}

	if fn == nil {
		LogD("skip %s", rpcmsg.Cmd)
		if rpcmsg.Cmd == "" {
			LogD("Skip req: %v || msg: (%d)%v", *rpcmsg, len(message), message)
		}
		return
	}

	rpcmsg.Sess = c.Sess
	ctx := MakeReqCtx(c, rpcmsg.Cmd, rpcmsg.Sess)

	//req_para := rpcmsg.Para
	//defer func() {
	//	LogD("req: %s rsp: %s err: %v", req_para, ret_data, err)
	//}()

	StatChg("recv "+rpcmsg.Cmd, 1)
	retmsg, err := fn(ctx, rpcmsg)
	if err == nil && retmsg != nil {
		retmsg.Sess = nil
		ret_data, err = jsoniter.Marshal(retmsg)
	}
	if c.Ch.Chmgr.DebugFlag > 0 {
		LogD("%s req: %s rsp: %s ", c.PeerAddr(), message, ret_data)
	}

	return
}

func (c *WebSock) HandleFunc(cmd string, fn interface{}) {
	if fn == nil {
		c.ctxHandleMap.Delete(cmd)
		return
	}
	c.ctxHandleMap.Store(cmd, fn)
}
func (c *WebSock) HandleFuncMap(mm map[string]ReqCtxHandler) {
	for cmd, nh := range mm {
		c.HandleFunc(cmd, nh)
	}
}
func (c *WebSock) DefHandleFunc(fn ReqCtxHandler) {
	c.defCtxHandler = fn
}

func (c *WebSock) RpcSend(rpcmsg *RpcMsg) (err error) {
	data, err := jsoniter.Marshal(rpcmsg)
	if err != nil || len(data) < 8 {
		err = NewError("Send failed: %s, len<11 or err:%v", string(data), err)
		return
	}
	return c.Ch.PushMsg(data)
}

func (c *WebSock) RpcPush(rpcmsg *RpcMsg) (err error) {
	rpcmsg.Chk = ""
	rpcmsg.Typ = RpcTypePush
	return c.RpcSend(rpcmsg)
}

func (c *WebSock) RpcRequest(rpcmsg *RpcMsg, wait_sec int64, pret interface{}) (err error) {
	defer func() {
		if errs := recover(); errs != nil {
			LogW("recover RecvRequest %s. err=%v", c.ConnInfo(), errs)
		}
	}()
	if c.Ws == nil {
		return errors.New("not ready")
	}

	sndid := atomic.AddUint64(&s_sndid, 1)
	rpcmsg.Chk = fmt.Sprintf("%s.%d", rpcmsg.Chk, sndid)
	rpcmsg.Typ = RpcTypeReq

	ch := make(chan *RpcMsg, 1)
	defer close(ch)

	s_sndmap.Store(rpcmsg.Chk, ch)
	defer s_sndmap.Delete(rpcmsg.Chk)

	StatChg("send "+rpcmsg.Cmd, 1)
	// c.Ch.PushMsg(rpcmsg)
	if err = c.RpcSend(rpcmsg); err != nil {
		return
	}

	t := time.NewTimer(time.Second * time.Duration(wait_sec))
	defer t.Stop()

	select {
	case rspmsg := <-ch:
		// rspmsg.Chk = rpcmsg.Chk
		if rspmsg.Err != nil {
			err = NewError("%s", rspmsg.Err.Error())
		} else if pret != nil {
			err = jsoniter.Unmarshal(rspmsg.Para, pret)
		}
		return
	// case x := <-time.After(time.Second * time.Duration(wait_sec)):
	case <-t.C:
		return errors.New("timeout")
	}

	return
}
