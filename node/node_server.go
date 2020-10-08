package node

// 1. handle websocket command.
// 2. auto
//
import (
	// "fmt"
	"net/http"
	"os"
	"strings"

	"stashware/oo"
	"stashware/types"

	"github.com/satori/go.uuid"
)

var (
	Version string
	Uuid    string = uuid.NewV4().String()

	debug_flag int64

	GNodeServer = &NodeServer{}
)

type NodeCfg = struct {
	types.ListenCfg
	PingSec   int64    `toml:"ping_sec"`
	PeerSec   int64    `toml:"peer_sec"`
	InitPeers []string `toml:"init_peers"`
	RejectIps []string `toml:"reject_ips"`
	Debug     int64    `toml:"debug"`
}

type NodeServer struct {
	chmgr      *oo.ChannelManager
	conf       NodeCfg
	ListenPort int64
	Chain      types.ChainIf
}

func NodeInit(gconf *oo.Config) (err error) {
	if err = gconf.SessDecode("node", &GNodeServer.conf); err != nil {
		return
	}
	conf := &GNodeServer.conf
	debug_flag = conf.Debug

	{
		if conf.PeerSec < int64(60) {
			conf.PeerSec = int64(60)
		}
		if conf.PingSec < int64(10) {
			conf.PingSec = int64(10)
		}
	}

	listens := strings.Split(conf.ListenCfg.Listen, ":")
	if len(listens) > 0 {
		GNodeServer.ListenPort = oo.Str2Int(listens[len(listens)-1])
	}
	if GNodeServer.ListenPort == 0 {
		GNodeServer.ListenPort = types.DEFAULT_SYNC_PORT
	}

	GNodeServer.chmgr = oo.NewChannelManager(64)
	GNodeServer.chmgr.DebugFlag = debug_flag

	return
}

func NodePostInit() (err error) {

	conf := &GNodeServer.conf

	if len(conf.Listen) > 0 {
		go func() {
			oo.LogD("Start node listen %v", conf.Listen)
			if err := ListenWsServer(conf.Listen, "", GNodeServer.chmgr, LinkOpen); err != nil {
				oo.LogW("Failed to listen node port.err %v ,conf %v", err, conf)
				os.Exit(1)
			}
		}()
	}

	if len(conf.SslListen) > 0 {
		go func() {
			oo.LogD("Start node ssl listen %v", conf.SslListen)
			if err := ListenWsServer(conf.SslListen, "", GNodeServer.chmgr, LinkOpen, conf.SslCert, conf.SslKey); err != nil {
				oo.LogW("Failed to listen node port. %v", conf)
				os.Exit(1)
			}
		}()
	}

	go NodeStartPeers()

	go NodeStartSync()

	go NodeSyncTxpool()

	return
}

func SetChain(chain types.ChainIf) {
	GNodeServer.Chain = chain
}

func SetVersion(ver string) {
	Version = ver
}

//listen one port, http or https
func ListenWsServer(addr string, webdir string, chmgr *oo.ChannelManager,
	open_fn func(*oo.WebSock) error, certkey ...string) error {

	var http_handler http.Handler
	if webdir != "" {
		http_handler = http.FileServer(http.Dir(webdir))
	}

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
		// if IsAddrReject(ip) {
		// 	oo.StatChg("reject addr", 1)
		// 	oo.LogD("reject %s", ip)
		// 	return
		// }

		if IsPeerConnected(ip) {
			oo.StatChg("double accept", 1)
			// oo.LogD("cancel double accept. %v", ip)
			return
		}
		oo.StatChg("active accept", 1)

		client.Sess = &oo.Session{
			Ip:     ip,
			Connid: client.Ch.GetSeq(),
		}
		// client.Data = r //for request uri / request header
		client.SetOpenHandler(open_fn)

		// client.Data = NodeEvent(types.NSEVENT_ACCEPTED, client)

		//not return
		client.RecvRequest()
		oo.StatChg("active accept", -1)
	})

	if len(certkey) < 2 || len(certkey[0]) == 0 {
		return http.ListenAndServe(addr, mux)
	} else {
		return http.ListenAndServeTLS(addr, certkey[0], certkey[1], mux)
	}
}
