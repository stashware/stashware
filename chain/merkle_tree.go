package chain

// type MerkerLeaf interface {
// 	Data() []byte
// }

// type MerkerLeafImpl struct {
// 	data []byte
// }

// func (this *MerkerLeafImpl) Data() []byte {
// 	return data
// }

type MerkerHash func([]byte) []byte

func GenerateMerkerRoot(hash MerkerHash, leaves [][]byte) []byte {
	if 0 == len(leaves) {
		return nil
	} else if 1 == len(leaves) {
		return leaves[0]
	}

	var nodes [][]byte
	for _, val := range leaves {
		nodes = append(nodes, val[:])
	}

	for len(nodes) > 1 {
		var (
			tmp [][]byte
			cnt int = len(nodes)
		)
		for i := 0; i < cnt; i += 2 {
			buf := nodes[i]
			if i+1 == cnt {
				buf = append(buf, nodes[i]...)
			} else {
				buf = append(buf, nodes[i+1]...)
			}
			tmp = append(tmp, hash(buf))
		}
		nodes = tmp
	}

	return nodes[0]
}
