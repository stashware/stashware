package wallet

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"stashware/oo"
)

type Wallet struct {
	pri_key *rsa.PrivateKey
	pub_key *rsa.PublicKey
}

func LoadWallet(private_key string) (ret *Wallet, err error) {
	buf := oo.HexDecodeString(private_key)

	key, err := x509.ParsePKCS1PrivateKey(buf)
	if nil != err {
		return
	}

	ret = &Wallet{
		pri_key: key,
		pub_key: &key.PublicKey,
	}

	return
}

func NewWallet() (ret *Wallet, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if nil != err {
		return
	}

	ret = &Wallet{
		pri_key: key,
		pub_key: &key.PublicKey,
	}

	return
}

func (this *Wallet) PrivateKey() string {
	buf := x509.MarshalPKCS1PrivateKey(this.pri_key)

	return oo.HexEncodeToString(buf)
}

func (this *Wallet) PublicKey() string {
	buf, _ := x509.MarshalPKIXPublicKey(this.pub_key)

	return oo.HexEncodeToString(buf)
}

func (this *Wallet) Address() string {
	buf, _ := x509.MarshalPKIXPublicKey(this.pub_key)

	return PublicKeyToAddress(buf)
}

func PublicKeyToAddress(pub []byte) string {
	buf := oo.Ripemd160(oo.Sha256(pub))

	buf = append([]byte{0x05}, buf...)

	checksum := func(payload []byte) []byte {
		firstSHA := oo.Sha256(payload)
		secondSHA := oo.Sha256(firstSHA[:])
		return secondSHA[:4]
	}

	buf = append(buf, checksum(buf)...)

	return oo.Base58EncodeString(buf)
}

func StringPublicKeyToAddress(pub_key string) string {
	return PublicKeyToAddress(oo.HexDecodeString(pub_key))
}

func Sign(pri_key string, msg []byte) (ret []byte, err error) {
	w, err := LoadWallet(pri_key)
	if nil != err {
		return
	}

	return rsa.SignPKCS1v15(rand.Reader, w.pri_key, crypto.SHA256, msg)
}

func VerifySignature(pub_key, signature string, msg []byte) (ret bool) {
	var (
		buf = oo.HexDecodeString(pub_key)
		sig = oo.HexDecodeString(signature)
	)

	pub, err := x509.ParsePKIXPublicKey(buf)
	if nil != err {
		return
	}

	err = rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, msg, sig)
	if nil != err {
		return
	}

	return true
}
