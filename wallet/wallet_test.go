package wallet

import (
	"testing"

	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	// "encoding/base64"
	"encoding/pem"

	"stashware/oo"
)

func TestStringPublicKeyToAddress(t *testing.T) {
	addr := StringPublicKeyToAddress("30820122300d06092a864886f70d01010105000382010f003082010a0282010100c9215878f3b9801d342dac0702287b27b170e0e005c908e9a5a4e491f3d8349dba57718de7422843525753d049dead294d999ac4a40e1f46c7995d1949bef770f242cd8957a0126dc375209bbfcec421ec45bcdcc6521d3764206500f6527335be27bf898f2452b0a4592f751c827dca94e2dc6385ed4a10a82ed5bbfd3cec395d345697cd487e2a7a26bd3f23a933d75dc3d59ac1170e46f4ef517205e72b16569578f2caeeefbb589fa98f86bfc4ff91dd2ada7e1d562897ced78061245be48914dcc28b2b91eb3806a3b9dfc2d73ef759be7a0ee2d8853e8c035b98613cb965cea2f6775dd34fc049c95b3f2f3757b0934c21b4f861e37f088d4a8aa9a7910203010001")

	t.Logf("%s", addr)
}

func TestLoadWallet(t *testing.T) {
	buf := oo.Base58DecodeString("KtDCWNWgpP9sYMqC8Z82nZdg9wcHR9iw2wJvQ6fqtq68kYrbLXpmyVZ5Gw55r58gG1UsyhSnhKY2qgoNDG8SxDeKce7cn436K995FAQc43mZkbH3DyEZr188ryEDZAnCVqft8cvnmS2FvqNcsXktNTgY3htAKXJDSkLR3n5fEexTW4Byi87gGkqEAjkDA1fH2WcYJaNskY5toJHh1ME4mxeBBBWLZNUWwkQ2shyS1NXFqwuX7FX1ExYUQSWC3TBGsdz6EHN9ykHFdhpF2v2Z1u6PigrUbKhSkufe8x7WZjyAvfATiShQFbgj4iQA4CCCT6nB1PwM6sBMSVLxALMea3p2JVAzjXQpSFQezBbgdyeUjJK57Kup7poAKTNdd6cSWwi2h3kHS9sedAgP5L2b6u9mkwK9Xma49pRo65MdgmvjKuKy1rBGnyHygqdwy3zx9b9bfVa5HeuwkopcbuKqR2qnKArc8ui38b8RYZ2VhRkR1EgDhaMbUc86fCSzHeEgjMkNtmFUomwR8kg5FgSfXwCRSrSFBjDgW6swmDX3yAg9YN6xtxm5uY8ERzk4fnRM3MGiWahLNEiC9iAtCeh46VzfKnp5YitzGhccGSMz5YghNbpBYMb9icWQ1V4zAJdYjUTrBDkUXvpdZWK2veXixjnqyNZsUnMTebcJz3Agh17fX7FECrPyarj3CZNE1Xcs9ZBcKV8qv5fGxyb2eGsPp5vagbV7N4mZhC2uwjv3kKicXbzWQbVEv4Lqgq1hkKSh9jsN8Mu8v1bfrFBuod5BPMZRRfVactY1dqSMbbpwNhQAn3sCZfT1rFbptDpTa3sfZR4Jjt6Pqi1U39EqrjmkwWvdZ6iQcLh7uVPPpMMe4tXCPtsVYC1ZvPQPEpTzcWyi2khSz5uL72UaG9P8wkq3qdhPTLMKrBAVCtz2Dj9CeKbfHR2D4WaNTQyqTKYFBvDkxycckLXGbTJyzEs5HQuT3HpMgPWHuKVxcvvCdRgKDF17WyQVrwbQUqPdaveUR1neQ4aCd3wqYnt1TmzoptYZWBSJh1nEuuUxv8yaMJbxf2WzguP74tthwx2x1ViRVEaJJA1TRByVajUvQNPpgknZNWBLaU5sEvJjCeV1wyT419Rmg86s6h8zGExBiyeCjZHAt65QBPdipJCCttRqSXBmtaq6rXKsxZUgRWUakD9xVgP83kfX83GfpRHhVZKnbudtVXHxRBeQCPKUNRNjqAuHk7kBu6Bcp7vvPoE8NUTmiGpZBrh3dJWP7vDfoN2t5RTeApLK91J67nmZ4NXq1T8F4194xqFQMrpEyR6dtihcQ32ciNQ8iVNEEgHzX7Z2y2uHE4NVYaU7niY4jZmLffC9gm8raQf8tCYDjzME4uACXcVK8CwYPQqGUAoU1fMY8antkUW4FehU5uYd9EeVoL6KJeeU6zf4xWPSNjeY44WJBhEUQ4CpacqbvbLz7oS8jRw3oYjA6Fp42Mue86ZLDj2GsDHJ1ji9kUcsCsMt3vTFzo3Q7wvmBCNfJf633A7smCqrdFpVo1SHnwjxACrN9tC2iTLeVMkYYgN1E1NPBw2e3maMH4npg5EshBaNCuZGTNhMM9bqNwvEUgmJkrjPstwkPLACw")
	w0, err := LoadWallet(oo.HexEncodeToString(buf))
	if nil != err {
		t.Fatal(err)
	}

	t.Log("pri-key", w0.PrivateKey())
	t.Log("pub-key", w0.PublicKey())
	t.Log("address", w0.Address()) // 3Nbzx3SsJubTgiMpzFLYiae2y3BYhprxRz
}

func TestWallet(t *testing.T) {
	w1, err := NewWallet()
	if nil != err {
		t.Fatal(err)
	}

	t.Logf("priv  == %s", w1.PrivateKey())
	t.Logf("pub   == %s", w1.PublicKey())
	t.Logf("addr  == %s", w1.Address())

	w2, err := LoadWallet(w1.PrivateKey())

	if w2.PublicKey() != w1.PublicKey() {
		t.Fatal(w2.PublicKey(), w1.PublicKey())
	}
	t.Log("pub1  == pub2")

	if w2.Address() != w1.Address() {
		t.Fatal(w2.Address(), w1.Address())
	}
	t.Log("addr1 == addr2")
}

func TestSignVerify(t *testing.T) {
	w1, err := NewWallet()
	if nil != err {
		t.Fatal(err)
	}

	raw := []byte("this is a msg")
	msg := oo.Sha256(raw)

	sign_msg, err := Sign(w1.PrivateKey(), msg)
	if nil != err {
		t.Fatal(err)
	}
	t.Logf("sign msg = %x", sign_msg)

	verify1 := VerifySignature(w1.PublicKey(), oo.HexEncodeToString(sign_msg), msg)

	t.Logf("verify1  = %v", verify1)

	w2, err := NewWallet()
	if nil != err {
		t.Fatal(err)
	}
	verify2 := VerifySignature(w2.PublicKey(), oo.HexEncodeToString(sign_msg), msg)

	t.Logf("verify2  = %v", verify2)
}
