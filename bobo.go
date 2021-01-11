package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/CortexFoundation/CortexTheseus/common"
	"github.com/CortexFoundation/CortexTheseus/common/hexutil"
	"github.com/CortexFoundation/CortexTheseus/common/math"
	"github.com/CortexFoundation/CortexTheseus/crypto"
	"github.com/CortexFoundation/CortexTheseus/crypto/secp256k1"

	"golang.org/x/crypto/sha3"

	badger "github.com/dgraph-io/badger/v2"
)

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

type Body struct {
	Timestamp int64  `json:"ts"`
	Addr      string `json:"addr"`
	//Text      string `json:"txt"`
}

type Msg struct {
	Timestamp int64  `json:"ts"`
	Text      string `json:"text"`
}

var (
	db            *badger.DB
	secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	testpri       = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
)

const (
	DigestLength = 32
	_FV_         = "_fv_"
	_FL_         = "_fl_"
	_PB_         = "_pb_"
	_ST_         = "_st_"
)

func main() {
	if bg, err := badger.Open(badger.DefaultOptions(".badger")); err == nil {
		defer bg.Close()
		db = bg
		mux := http.NewServeMux()
		mux.HandleFunc("/", handler)
		http.ListenAndServe("127.0.0.1:8080", mux)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v", r.Method, r.URL)
	res := "OK"

	uri := strings.ToLower(r.URL.Path)
	u := strings.Split(uri, "/")
	if len(u) < 2 {
		fmt.Fprintf(w, "Invalid URL")
		return
	}

	addr, category := u[len(u)-1], u[len(u)-2]
	if !common.IsHexAddress(addr) {
		fmt.Fprintf(w, "Invalid infohash format")
		return
	}
	q := r.URL.Query()
	switch r.Method {
	case http.MethodGet:
		switch category {
		case "user":
			res = UserDetails(uri)
		case "favor":
			res = FavorList(uri)
		case "favored":
			res = FavoredList(addr)
		case "follow":
			res = FollowList(uri)
		case "followed":
			res = FollowedList(addr)
		case "msg":
			res = MsgList(addr)
		default:
			res = "Method not found"
		}
	case http.MethodPost:
		if reqBody, err := ioutil.ReadAll(r.Body); err == nil {
			var body Body
			var to string
			if category != "work" {
				if len(reqBody) > 0 {
					if err := json.Unmarshal(reqBody, &body); err != nil {
						log.Printf("%v", err)
						res = "Invalid json"
						break
					}
				}

				to = strings.ToLower(body.Addr)
				if len(to) > 0 && !common.IsHexAddress(to) {
					res = "Invalid addr format"
					break
				}

				if !Verify(string(reqBody), addr, q.Get("sig"), body.Timestamp) {
					res = "Invalid signature"
					break
				}
			}

			switch category {
			case "user":
				if err := UserCreate(uri, string(reqBody)); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			case "favor":
				if err := Favor(uri, to); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			case "follow":
				if err := Follow(uri, to); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			case "artist":
				if err := PubWork(uri, to); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			case "work":
				status := "changed"
				if len(q.Get("st")) > 0 {
					status = q.Get("st")
				}
				if err := UpdateWork(uri, status); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			default:
				res = "Method not found"
			}
		}
	case http.MethodDelete:
		if reqBody, err := ioutil.ReadAll(r.Body); err == nil {
			var body Body
			if err := json.Unmarshal(reqBody, &body); err != nil {
				log.Printf("%v", err)
				res = fmt.Sprintf("%v", err)
				break
			}

			to := strings.ToLower(body.Addr)
			if len(to) > 0 && !common.IsHexAddress(to) {
				res = "Invalid addr format"
				break
			}

			if !Verify(string(reqBody), addr, q.Get("sig"), body.Timestamp) {
				res = "Invalid signature"
				break
			}

			switch category {
			case "favor":
				if err := Unfavor(uri, to); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			case "follow":
				if err := Unfollow(uri, to); err != nil {
					res = fmt.Sprintf("%v", err)
				}
			default:
				res = "Method not found"
			}
		}
	default:
		res = "INVALID REQUEST TYPE"
	}
	fmt.Fprintf(w, res)
}

func Unfavor(uri, to string) error {
	return del(uri + _FV_ + to)
}

func Unfollow(uri, to string) error {
	return del(uri + _FL_ + to)
}

func UserCreate(uri, v string) error {
	return set(uri, v)
}

func Favor(uri, to string) error {
	return set(uri+_FV_+to, to)
}

func Follow(uri, to string) error {
	return set(uri+_FL_+to, to)
}

func PubWork(uri, to string) error {
	m := Msg{Timestamp: time.Now().Unix(), Text: to}
	if res, err := json.Marshal(m); err == nil {
		go setTTL(uri+_PB_+to, string(res), 24*time.Hour)
	}
	return nil
}

func UpdateWork(uri, status string) error {
	m := Msg{Timestamp: time.Now().Unix(), Text: status}
	if res, err := json.Marshal(m); err == nil {
		go setTTL(uri+_ST_+status, string(res), 24*time.Hour)
	}
	return nil
}

func UserDetails(k string) string {
	return get(k)
}

func Verify(msg, addr, sig string, timestamp int64) bool {
	if len(msg) == 0 || len(addr) == 0 || len(sig) == 0 || timestamp == 0 {
		return false
	}

	if time.Now().Unix()-int64(30) > timestamp {
		//return errors.New("Signature expired")
		//TODO
		//return false
	}

	if time.Now().Unix()+int64(15) < timestamp {
		//return errors.New("Signature disallowed future")
		//TODO
		//return false
	}

	sig_, _ := SignHex(msg, testpri)
	log.Printf("[signature] : %s", hexutil.Encode(sig_[:]))

	m := Keccak256([]byte(msg))
	s := hexutil.MustDecode(sig)

	if len(m) == 0 || len(s) == 0 {
		return false
	}

	recoveredPub, err := Ecrecover(m, s)
	if err != nil {
		return false
	}

	pubKey, _ := UnmarshalPubkey(recoveredPub)
	recoveredAddr := PubkeyToAddress(*pubKey)
	if common.HexToAddress(addr) != recoveredAddr {
		log.Printf("Address mismatch: want: %v have: %v\n", addr, recoveredAddr.Hex())
		return false
	}

	if !VerifySignature(recoveredPub, m, s[:len(s)-1]) {
		return false
	}
	return true
}

func FavorList(k string) string {
	res, _ := json.Marshal(prefix(k))
	return string(res)
}

func FollowList(k string) string {
	res, _ := json.Marshal(prefix(k))
	return string(res)
}

func FollowedList(k string) string {
	k = _FL_ + k
	followers := suffix(k)

	var tmp []string
	for _, f := range followers {
		vs := strings.Split(string(f), _FL_)
		fs := strings.Split(vs[0], "/")
		if len(fs) > 0 {
			tmp = append(tmp, fs[len(fs)-1])
		}

	}
	res, _ := json.Marshal(tmp)
	return string(res)
}

func MsgList(k string) string {
	fls := prefix("/follow/" + k)
	var tmp []Msg
	for _, fl := range fls {
		log.Println("follow : " + fl)
		msgs := prefix("/artist/" + fl)
		for _, m := range msgs {
			log.Println("artist : [" + fl + "] has published a new work [" + m + "]")
			var mm Msg
			if err := json.Unmarshal([]byte(m), &mm); err == nil {
				mm.Text = "Artist " + fl + " published " + mm.Text
				tmp = append(tmp, mm)
			}
		}
	}

	fvs := prefix("/favor/" + k)
	for _, fv := range fvs {
		log.Println("favor : " + fv)
		msgs := prefix("/work/" + fv + _ST_)
		for _, m := range msgs {
			log.Println("Work : [" + fv + "] status update [" + m + "]")
			var mm Msg
			if err := json.Unmarshal([]byte(m), &mm); err == nil {
				mm.Text = "Work " + fv + " is " + mm.Text
				tmp = append(tmp, mm)
			}
		}
	}

	res, _ := json.Marshal(tmp)
	return string(res)
}

func FavoredList(k string) string {
	k = _FV_ + k
	favs := suffix(k)

	var tmp []string
	for _, f := range favs {
		vs := strings.Split(string(f), _FV_)
		fs := strings.Split(vs[0], "/")
		if len(fs) > 0 {
			tmp = append(tmp, fs[len(fs)-1])
		}
	}
	res, _ := json.Marshal(tmp)
	return string(res)
}

func get(k string) (val string) {
	if len(k) == 0 {
		return
	}
	db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get([]byte(k)); err == nil {
			//if val, err := item.ValueCopy(nil); err == nil {
			//	v = string(val)
			//}

			item.Value(func(v []byte) error {
				//	fmt.Printf("key=%s, value=%s\n", k, v)
				val = string(v)
				return nil
			})
		}
		return nil
	})
	return
}

func set(k, v string) (err error) {
	if len(k) == 0 || len(v) == 0 {
		return
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(k), []byte(v))
	})
	return
}

func setTTL(k, v string, expire time.Duration) (err error) {
	if len(k) == 0 || len(v) == 0 || expire == 0 {
		return
	}
	log.Println("TTL " + k + ", " + v)
	err = db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(k), []byte(v)).WithTTL(expire)
		return txn.SetEntry(e)
	})
	return
}

func del(k string) (err error) {
	if len(k) == 0 {
		return
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(k))
	})
	return
}

// k="/favor/0x2a2a0667f9cbf4055e48eaf0d5b40304b8822184"
func prefix(k string) (res []string) {
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(k)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			//k := item.Key()
			item.Value(func(v []byte) error {
				//fmt.Printf("key=%s, value=%s\n", k, v)
				res = append(res, string(v))
				return nil
			})
			//if val, err := item.ValueCopy(nil); err == nil {
			//	res = append(res, string(val))
			//}

		}
		return nil
	})

	return
}

func scan() (res []string) {
	db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			//			k := item.Key()
			err := item.Value(func(v []byte) error {
				//				fmt.Printf("key=%s, value=%s\n", k, v)
				res = append(res, string(v))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func suffix(suf string) (res []string) {
	db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if strings.HasSuffix(string(k), suf) {
				res = append(res, string(k))
			}
		}
		return nil
	})
	return
}

func sequence(key string) {
	seq, _ := db.GetSequence([]byte(key), 1000)
	defer seq.Release()
	num, _ := seq.Next()
	log.Printf("seq %v", num)
}

func VerifySignature(pubkey, hash, signature []byte) bool {
	return secp256k1.VerifySignature(pubkey, hash, signature)
}

func EcRecover(data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Cortex signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1
	hash, _ := SignHash(data)
	rpk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

func SignHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Cortex Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg)), msg
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := sha3.NewLegacyKeccak256().(KeccakState)
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

func UnmarshalPubkey(pub []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(S256(), pub)
	if x == nil {
		return nil, errors.New("invalid key")
	}
	return &ecdsa.PublicKey{Curve: S256(), X: x, Y: y}, nil
}

func S256() elliptic.Curve {
	return secp256k1.S256()
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}

func SignHex(msg string, pri string) (sig []byte, err error) {
	k0, _ := HexToECDSA(pri)
	msg0 := Keccak256([]byte(msg))
	return Sign(msg0, k0)
}

func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != DigestLength {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	return secp256k1.Sign(hash, seckey)
}

func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if byteErr, ok := err.(hex.InvalidByteError); ok {
		return nil, fmt.Errorf("invalid hex character %q in private key", byte(byteErr))
	} else if err != nil {
		return nil, errors.New("invalid hex data for private key")
	}
	return ToECDSA(b)
}

func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	return toECDSA(d, true)
}

func toECDSA(d []byte, strict bool) (*ecdsa.PrivateKey, error) {
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	if strict && 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	// The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}
