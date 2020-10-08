package web_wallet

import (
	"github.com/gobuffalo/packr/v2"
	// _ "github.com/gobuffalo/packr/v2/packr2" //just for packr2 binary
	"net/http"
	"os"
	"stashware/node"
	"stashware/oo"
	"stashware/types"
	"strings"
)

type WebWalletCfg = struct {
	types.ListenCfg
	Debug int64 `toml:"debug"`
}

type WebWalletServer struct {
	conf WebWalletCfg
}

var GWebWalletServer = &WebWalletServer{}
var debug_flag int64

func WebWalletInit(gconf *oo.Config) (err error) {
	if err = gconf.SessDecode("web_wallet", &GWebWalletServer.conf); err != nil {
		return
	}
	conf := &GWebWalletServer.conf
	debug_flag = conf.Debug

	return
}

func WebWalletPostInit() (err error) {
	conf := &GWebWalletServer.conf

	box := packr.New("wallet_h5", "../../app/wallet_h5")

	handler := http.FileServer(box)

	chmgr := oo.NewChannelManager(64)
	chmgr.DebugFlag = debug_flag

	if len(conf.Listen) > 0 {
		go func() {
			if err = WalletServer(conf.Listen, handler, chmgr, WalletOpen); err != nil {
				oo.LogW("Failed to listen wallet port %s, err=%v", conf.Listen, err)
				os.Exit(1)
			}
		}()
	}

	if len(conf.SslListen) > 0 {
		go func() {
			if err = WalletServer(conf.SslListen, handler, chmgr, WalletOpen, conf.SslCert, conf.SslKey); err != nil {
				oo.LogW("Failed to listen wallet port %s, err=%v", conf.Listen, err)
				os.Exit(1)
			}
		}()
	}
	return
}

func WalletOpen(ws *oo.WebSock) (err error) {
	ws.OpenCoroutineFlag()
	ws.SetCloseHandler(func(c *oo.WebSock) {
		oo.LogD("wallet conn close : %v.", c.ConnInfo())
	})
	// ws.SetIntervalHandler(AutoPing)
	ws.HandleFuncMap(node.GNodeHandlers)
	oo.LogD("wallet conn open: %v", ws.ConnInfo())
	// ws.HandleFunc(types.CMD_INFO, OnReqInfo)
	return
}

func WalletServer(addr string, http_handler http.Handler, chmgr *oo.ChannelManager,
	open_fn func(*oo.WebSock) error, certkey ...string) error {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, exists := r.Header["Upgrade"]; !exists {
			if http_handler != nil {
				http_handler.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}
		if open_fn == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		client, err := oo.InitWebSock(w, r, chmgr, oo.RecvRpcFn)
		if err != nil {
			oo.LogD("Failed to init ws. err:%v", err)
			return
		}
		defer client.Close()

		ip := r.RemoteAddr
		if fw := r.Header.Get("X-Forwarded-For"); fw != "" {
			ip = fw
			if ips := strings.Split(fw, ", "); len(ips) > 1 {
				ip = ips[0]
			}
		}

		client.Sess = &oo.Session{
			Ip:     ip,
			Connid: client.Ch.GetSeq(),
		}
		client.Data = r //for request uri / request header
		client.SetOpenHandler(open_fn)

		oo.LogD("newconn %s", ip)
		// save peer
		// atomic.AddInt64(&peer_count, 1)
		// peer_addr.Store(r.URL.String(), 1)
		// defer func() {
		// 	atomic.AddInt64(&peer_count, -1)
		// 	peer_addr.Delete(r.URL.String())
		// }()

		//not return
		client.RecvRequest()
	})

	if len(certkey) < 2 || len(certkey[0]) == 0 {
		return http.ListenAndServe(addr, mux)
	} else {
		return http.ListenAndServeTLS(addr, certkey[0], certkey[1], mux)
	}
}
