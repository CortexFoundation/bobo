package main

import (
	"fmt"
	badger "github.com/dgraph-io/badger/v2"
	"log"
	"net/http"
)

var db *badger.DB

func main() {
	opts := badger.DefaultOptions("/tmp/badger")
	b, err := badger.Open(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()
	db = b

	fmt.Println("Badger started")

	http.HandleFunc("/", Handler)

	http.ListenAndServe(":8080", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	method := r.URL.Path[1:]
	q := r.URL.Query()

	var res = "suc"
	switch method {
	case "get":
		res = Get(q.Get("k"))
	case "set":
		err := Set(q.Get("k"), q.Get("v"))
		if err != nil {
			res = "failed"
		}
	default:
		res = Default()
	}

	fmt.Fprintf(w, res)
}

func Get(k string) string {
	fmt.Println("Do get key=" + k)

	//TODO

	return get(k)
}

func Set(k, v string) error {
	fmt.Println("Do set key=" + k + ", value=" + v)

	//TODO

	return set(k, v)
}

func Default() string {
	return "method not found"
}

func set(k, v string) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(k), []byte(v))
		return err
	})
	return err
}

func get(k string) string {
	var v string
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		v = string(val)
		fmt.Printf("The answer is: %s\n", val)
		return nil
	})

	if err != nil {
		return ""
	}
	return v
}
