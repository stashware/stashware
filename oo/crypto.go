package oo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

type AesCrypt struct {
	Passwd   string `toml:"passwd,omitzero"`
	Ivstring string `toml:"ivstring,omitzero"`
}

func Md5(data []byte) []byte {
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	return md5Ctx.Sum(nil)
}

func Md5Str(data []byte) string {
	return hex.EncodeToString(Md5(data))
}

func Sha1(data []byte) []byte {
	h := sha1.New()
	return h.Sum(data)
}

func Sha1Str(data []byte) string {
	return hex.EncodeToString(Sha1(data))
}

func HmacSha1(data []byte, key []byte) []byte {
	h := hmac.New(sha1.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func HmacSha1Str(data []byte, key []byte) string {
	return hex.EncodeToString(HmacSha1(data, key))
}

func HmacSha256(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func HmacSha256Str(data []byte, key []byte) string {
	return hex.EncodeToString(HmacSha256(data, key))
}

func HashTime33(str string) int {
	bStr := []byte(str)
	bLen := len(bStr)
	hash := 5381

	for i := 0; i < bLen; i++ {
		hash += ((hash << 5) & 0x7FFFFFF) + int(bStr[i])
	}
	return hash & 0x7FFFFFF
}

func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// https://en.bitcoin.it/wiki/Base58Check_encoding#Version_bytes
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Base58Decode decodes Base58-encoded data
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(b58Alphabet, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	if len(input) > 0 && input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

func Base58EncodeString(buf []byte) string {
	if 0 == len(buf) {
		return ""
	}
	return fmt.Sprintf("%s", Base58Encode(buf))
}

func Base58DecodeString(str string) []byte {
	return Base58Decode(Str2Bytes(str))
}

func HexDecodeString(str string) []byte {
	buf, _ := hex.DecodeString(str)

	return buf
}

func HexEncodeToString(buf []byte) string {
	return hex.EncodeToString(buf)
}

func HexDecStringPad32(str string) []byte {
	buf := HexDecodeString(str)

	buf = paddedAppend(32, buf)

	return buf
}

func HexDecStringPad32OrNil(str string) []byte {
	buf := HexDecodeString(str)

	if len(buf) > 0 {
		buf = paddedAppend(32, buf)
	}

	return buf
}

func HexEncToStringPad32OrNil(buf []byte) string {
	if len(buf) > 0 {
		buf = paddedAppend(32, buf)
	}

	return hex.EncodeToString(buf)
}

func HexEncToStringPad32(buf []byte) string {
	buf = paddedAppend(32, buf)

	return hex.EncodeToString(buf)
}

func paddedAppend(size int, src []byte) []byte {
	var (
		pre []byte
	)
	for i := 0; i < int(size)-len(src); i++ {
		pre = append(pre, 0)
	}
	return append(pre, src...)
}

// aes
func AesEncrypt(passwd, ctx []byte) []byte {
	if len(passwd) > 0 {
		block, _ := aes.NewCipher(passwd)
		blockSize := block.BlockSize()
		padding := blockSize - len(ctx)%blockSize
		paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
		ctx = append(ctx, paddingText...)
		blockMode := cipher.NewCBCEncrypter(block, passwd[:blockSize])
		plainText := make([]byte, len(ctx))
		blockMode.CryptBlocks(plainText, ctx)
		return plainText
	}
	return ctx
}

func AesDecrypt(passwd, ctx []byte) []byte {
	if len(passwd) > 0 {
		block, _ := aes.NewCipher(passwd)
		blockSize := block.BlockSize()
		blockMode := cipher.NewCBCDecrypter(block, passwd[:blockSize])
		plainText := make([]byte, len(ctx))
		blockMode.CryptBlocks(plainText, ctx)
		tLen := len(plainText)
		plainText = plainText[:tLen-int(plainText[tLen-1])]
		return plainText
	}
	return ctx
}

func Sha256(buf []byte) []byte {
	ret := sha256.Sum256(buf)

	return ret[:]
}

func Ripemd160(buf []byte) []byte {
	h := ripemd160.New()

	_, _ = h.Write(buf[:])

	ret := h.Sum(nil)

	return ret[:]
}
