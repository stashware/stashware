package rest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestMux(t *testing.T) {
	var Fns = map[string]func(http.ResponseWriter, *http.Request){
		"/info": func(http.ResponseWriter, *http.Request) {
			fmt.Printf("on info\n")
		},
		"/peers": func(http.ResponseWriter, *http.Request) {
			fmt.Printf("on peers\n")
		},
		"/{txid:.{64}}": func(w http.ResponseWriter, r *http.Request) {
			params := mux.Vars(r)
			fmt.Printf("on txid=%s\n", params["txid"])
		},
	}

	m := mux.NewRouter()
	m.HandleFunc("/", func(http.ResponseWriter, *http.Request) {
		fmt.Printf("on root\n")
	})

	for cmd, fn := range Fns {
		m.HandleFunc(cmd, fn)
	}
	http.ListenAndServe(":8888", m)
}
