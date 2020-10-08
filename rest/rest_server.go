package rest

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"stashware/miner"
	"stashware/node"
	"stashware/oo"
	"stashware/types"
	"strconv"
	// "strings"
)

type GatewayCfg = struct {
	types.ListenCfg
	Debug int64 `toml:"debug"`
}

type RestServer struct {
	conf GatewayCfg
}

var RestPath2SyncCmd = map[string]func(http.ResponseWriter, *http.Request){
	"/info":                          InfoRequest,  //types.CMD_INFO,
	"/peers":                         PeersRequest, //types.CMD_PEERS,
	"/debug":                         DebugRequest,
	"/tx/pending":                    TxPendingRequest,    //
	"/tx/{txid:.{64}}":               TxByIdRequest,       //types.CMD_TX_BY_ID,
	"/txs/{address:.{34}}":           TxsAddressRequest,   //
	"/txs/{address:.{34}}/{h1}":      TxsAddressRequest,   //
	"/txs/{address:.{34}}/{h1}/{h2}": TxsAddressRequest,   //
	"/{txid:.{64}}":                  TxDataRequest,       //types.CMD_TX_DATA,
	"/{txid:.{64}}.{extension}":      TxDataRequest,       //types.CMD_TX_DATA,
	"/price/{bytes}/{target:.{34}}":  PriceRequest,        //types.CMD_PRICE,
	"/tx":                            SubmitTxRequest,     //types.CMD_SUBMIT_TX,
	"/wallet/new":                    WalletNew,           //types.CMD_WALLET_NEW,
	"/wallet/{address:.{34}}":        WalletByAddrRequest, //types.CMD_WALLET_BALANCE,
	"/block/hash/{block_id:.{64}}":   GetBlockRequest,     //types.CMD_BLOCK_BY_ID,
	"/block/height/{height}":         GetBlockRequest,     //types.CMD_BLOCK_BY_ID,
	// "/tx/{id}/{field}":          TxFieldRequest,      //types.CMD_TX_FIELD,
	"/mining": GetMiningInfo,
}

var GRestServer = &RestServer{}
var debug_flag int64

func RestInit(gconf *oo.Config) (err error) {
	if err = gconf.SessDecode("gateway", &GRestServer.conf); err != nil {
		return
	}
	conf := &GRestServer.conf
	debug_flag = conf.Debug

	return
}

func RestPostInit() (err error) {
	conf := &GRestServer.conf

	// mux := http.NewServeMux()
	mux := mux.NewRouter()
	mux.HandleFunc("/", RestRequest)
	for k, fn := range RestPath2SyncCmd {
		mux.HandleFunc(k, fn)
	}

	if len(conf.Listen) > 0 {
		go func() {
			if err := http.ListenAndServe(conf.Listen, mux); err != nil {
				oo.LogW("Failed to start rest. err=%v", err)
				os.Exit(1)
			}
		}()
	}

	if len(conf.SslListen) > 0 {
		go func() {
			if err := http.ListenAndServeTLS(conf.Listen, conf.SslCert, conf.SslKey, mux); err != nil {
				oo.LogW("Failed to start rest. err=%v", err)
				os.Exit(1)
			}
		}()
	}
	return
}

func RestRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	w.Header().Set("Content-Type", "application/json")

	reqmsg := &oo.RpcMsg{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil || oo.JsonUnmarshal(data, reqmsg) != nil {

	}

}

func InfoRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	reqmsg := oo.PackRequest(types.CMD_INFO, nil, nil)
	rspmsg, err := node.OnReqInfo(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func PeersRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	reqmsg := oo.PackRequest(types.CMD_PEERS, nil, nil)
	rspmsg, err := node.OnReqPeers(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v, rspmsg=%v", err, rspmsg)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func DebugRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	reqmsg := oo.PackRequest(types.CMD_DEBUG, nil, nil)
	rspmsg, err := node.OnReqDebug(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v, rspmsg=%v", err, rspmsg)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}
func TxPendingRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	reqmsg := oo.PackRequest(types.CMD_TX_PENDING, nil, nil)
	rspmsg, err := node.OnReqTxPending(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func TxByIdRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdTxByIdReq{
		Txid: params["txid"],
	}
	reqmsg := oo.PackRequest(types.CMD_TX_BY_ID, para, nil)

	rspmsg, err := node.OnReqTxById(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func TxsAddressRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdGetAddrTxsReq{
		Address: params["address"],
	}

	if "" != params["h1"] {
		num, err := strconv.ParseInt(params["h1"], 10, 64)
		if nil != err {
			oo.LogD("GetBlockRequest parse h1[%s] err %v", params["h1"], err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		para.Height1 = &num
	}

	if "" != params["h2"] {
		num, err := strconv.ParseInt(params["h2"], 10, 64)
		if nil != err {
			oo.LogD("GetBlockRequest parse h2[%s] err %v", params["h2"], err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		para.Height2 = &num
	}

	reqmsg := oo.PackRequest(types.CMD_GET_ADDR_TXS, para, nil)

	rspmsg, err := node.OnReqGetAddrTxs(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func TxDataRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdTxDataReq{
		Txid:      params["txid"],
		Extension: params["extension"],
	}

	content_type := mime.TypeByExtension("." + para.Extension)
	if "" != content_type {
		w.Header().Set("Content-Type", content_type)

	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		oo.LogD("TxDataRequest error extension[%s]", para.Extension)
	}

	reqmsg := oo.PackRequest(types.CMD_TX_DATA, para, nil)
	rspmsg, err := node.OnReqTxData(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	var data string
	oo.JsonUnmarshal(rspmsg.Para, &data)

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Write(oo.HexDecodeString(data))
	}
}

func PriceRequest(w http.ResponseWriter, r *http.Request) {
	var err error

	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdPriceReq{
		Bytes:  oo.Str2Int(params["bytes"]),
		Target: params["target"],
	}
	reqmsg := oo.PackRequest(types.CMD_PRICE, para, nil)
	rspmsg, err := node.OnReqPrice(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func SubmitTxRequest(w http.ResponseWriter, r *http.Request) {
	if "POST" != r.Method {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer r.Body.Close()

	buf, err := ioutil.ReadAll(r.Body)
	if nil != err {
		oo.LogD("SubmitTxRequest ioutil.ReadAll %v", err)
		return
	}

	reqmsg := &oo.RpcMsg{
		Cmd:  types.CMD_SUBMIT_TX,
		Typ:  oo.RpcTypeReq,
		Para: buf,
		Sess: nil,
	}

	rspmsg, err := node.OnReqSubmitTx(&oo.ReqCtx{}, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}

	return
}

func WalletNew(w http.ResponseWriter, r *http.Request) {
	reqmsg := oo.PackRequest(types.CMD_WALLET_NEW, nil, nil)

	rspmsg, err := node.OnReqWalletNew(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func WalletByAddrRequest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdWalletByAddrReq{
		Address: params["address"],
	}
	reqmsg := oo.PackRequest(types.CMD_WALLET_BY_ADDR, para, nil)

	rspmsg, err := node.OnReqWalletByAddr(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func GetBlockRequest(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	oo.LogD("on http request params : %v", params)

	para := types.CmdGetBlockReq{
		BlockId: params["block_id"],
		// BlockHeight: params["height"],
	}
	if "" != params["height"] {
		num, err := strconv.ParseInt(params["height"], 10, 64)
		if nil != err {
			oo.LogD("GetBlockRequest parse height[%s] err %v", params["height"], err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		para.BlockHeight = num
	}

	reqmsg := oo.PackRequest(types.CMD_GET_BLOCK, para, nil)

	rspmsg, err := node.OnReqGetBlock(nil, reqmsg)
	if err != nil || (nil != rspmsg && nil != rspmsg.Err) {
		oo.LogD("badgw err=%v", err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	if rspmsg != nil && len(rspmsg.Para) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write(rspmsg.Para)
	}
}

func GetMiningInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	info := miner.GetMiningInfo()
	data := oo.JsonData(*info)
	w.Write(data)
}
