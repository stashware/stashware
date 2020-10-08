package node

import (
	"testing"
)

func TestSelectSubset(t *testing.T) {
	peers := []string{
		"aa",
		"bb",
		"cc",
	}

	Ip2Score.Store(peers[1], true)

	peers = SelectSubset(peers)
	t.Logf("== %v", peers)
}

func TestSelectSyncNodes(t *testing.T) {

}

func TestNodeTopN(t *testing.T) {

}
