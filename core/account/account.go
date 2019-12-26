package account

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/hacash/bitcoin/address/address"
	"github.com/hacash/bitcoin/address/btcec"
	"github.com/hacash/blockmint/block/fields"
	"regexp"
	"strconv"
)

type Account struct {
	AddressReadable fields.AddressReadable
	Address         fields.Address
	PublicKey       []byte
	PrivateKey      []byte
	Private         *btcec.PrivateKey
}

func GetAccountByPriviteKeyString(hexstr string) (*Account, error) {
	byte, e1 := hex.DecodeString(hexstr)
	if e1 != nil {
		return nil, e1
	}
	return GetAccountByPriviteKey(byte)
}

func GetAccountByPriviteKey(byte []byte) (*Account, error) {
	privt, e2 := btcec.ToECDSA(byte)
	if e2 != nil {
		return nil, e2
	}
	private := btcec.PrivateKey(*privt)
	return genAccountByPrivateKey(private), nil
}

func CreateAccountByPassword(password string) *Account {
	digest := sha256.Sum256([]byte(password))
	privite, _ := btcec.PrivKeyFromBytes(btcec.S256(), digest[:])
	return genAccountByPrivateKey(*privite)
}

func CreateNewAccount() *Account {
	digest := make([]byte, 32)
	rand.Read(digest)
	privite, _ := btcec.PrivKeyFromBytes(btcec.S256(), digest)
	return genAccountByPrivateKey(*privite)
}

func genAccountByPrivateKey(private btcec.PrivateKey) *Account {
	compressedPublic := private.PubKey().SerializeCompressed()
	addr := address.NewAddressFromPublicKey([]byte{0}, compressedPublic)
	readable := address.NewAddressReadableFromAddress(addr)
	return &Account{
		AddressReadable: fields.AddressReadable(readable),
		Address:         addr,
		PublicKey:       compressedPublic,    // 压缩公钥
		PrivateKey:      private.Serialize(), // 私钥
		Private:         &private,
	}
}

func FindNiceAccounts(basestr string, prenum int, start int64) {

	for i := start; ; i++ {
		password := basestr + strconv.FormatInt(i, 10)
		addr := CreateAccountByPassword(password)
		rdble := addr.AddressReadable
		prefix := regexp.MustCompile(`^[1-9]+`).FindAllString(string(rdble), -1)[0]
		if len(prefix) >= prenum {
			tailstr := string([]byte(rdble)[len(prefix):])
			//fmt.Println(tailstr)
			havnum := regexp.MustCompile(`[1-9]+`).FindAllString(tailstr, -1)
			if len(havnum) == 0 {
				//fmt.Println(password)
				//fmt.Println(i, "=", hex.EncodeToString(addr.PrivateKey), ":", rdble)
				fmt.Println(password, rdble)
			}
		}
	}

}

func FindNiceAccounts2(basestr string, start uint64, total uint64) {

	var n uint64 = 0
	for i := start; ; i++ {
		password := basestr + strconv.FormatUint(i, 10)
		addr := CreateAccountByPassword(password)
		rdble := addr.AddressReadable
		macth := regexp.MustCompile(`(^1[1-9]+$)|(^1[A-Za-z]+$)`).FindAllString(string(rdble), -1)
		if len(macth) > 0 && len(macth[0]) > 0 {
			//fmt.Println(password)
			fmt.Println(n, " ", i, " ", rdble)
			n++
			if n >= total {
				break
			}
			//fmt.Println(password, rdble)
		}
	}

}
