package db

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/base/agent/log"

	"github.com/subutai-io/gorjun/config"
)

var (
	bucket = []byte("MyBucket")
	search = []byte("SearchIndex")
	users  = []byte("Users")
	tokens = []byte("Tokens")
	authid = []byte("AuthID")
	db     = initdb()
)

func initdb() *bolt.DB {
	db, err := bolt.Open("my.db", 0600, nil)
	log.Check(log.FatalLevel, "Openning DB: my.db", err)

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucket, search, users, tokens, authid} {
			_, err := tx.CreateBucketIfNotExists(b)
			log.Check(log.FatalLevel, "Creating bucket: "+string(b), err)
		}
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction", err)
	return db
}

func Write(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "public"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()

		b, _ := tx.Bucket(users).CreateBucketIfNotExists([]byte(owner))
		b, _ = b.CreateBucketIfNotExists([]byte("files"))
		b.Put([]byte(key), []byte(value))

		b, _ = tx.Bucket(bucket).CreateBucketIfNotExists([]byte(key))
		b.Put([]byte("date"), now)
		b.Put([]byte("name"), []byte(value))
		b.Put([]byte("owner"), []byte(owner))

		f, err := os.Open(config.Filepath + key)
		if !log.Check(log.WarnLevel, "Opening file "+config.Filepath+key, err) {
			fi, _ := f.Stat()
			f.Close()
			b.Put([]byte("size"), []byte(fmt.Sprint(fi.Size())))
		}

		for i, _ := range options {
			for k, v := range options[i] {
				b.Put([]byte(k), []byte(v))
			}
		}

		b, _ = tx.Bucket(search).CreateBucketIfNotExists([]byte(value))
		b.Put(now, []byte(key))

		return nil
	})
	log.Check(log.WarnLevel, "Writing data to db", err)
}

func Delete(key string) {
	db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucket).DeleteBucket([]byte(key))
	})
}

func Read(key string) (val string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(key)); b != nil {
			if value := b.Get([]byte("name")); value != nil {
				val = string(value)
			}
		}
		return nil
	})
	return val
}

func List() (list map[string]string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		b.ForEach(func(k, v []byte) error {
			if value := b.Bucket(k).Get([]byte("name")); value != nil {
				list[string(k)] = string(value)
			}
			return nil
		})
		return nil
	})
	return list
}

func Info(hash string) (list map[string]string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(bucket).Bucket([]byte(hash)); b != nil {
			b.ForEach(func(k, v []byte) error {
				list[string(k)] = string(v)
				return nil
			})
		}
		return nil
	})
	return list
}

func Close() {
	db.Close()
}

func Search(query string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(search)
		c := b.Cursor()
		for k, _ := c.Seek([]byte(query)); len(k) > 0 && bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
			b.Bucket(k).ForEach(func(kk, vv []byte) error {
				list[string(vv)] = string(k)
				return nil
			})
		}
		return nil
	})
	return list
}

func LastHash(name string) (hash string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(search).Bucket([]byte(name)); b != nil {
			_, v := b.Cursor().Last()
			hash = string(v)
		}
		return nil
	})
	return hash
}

func RegisterUser(name, key []byte) {
	db.Update(func(tx *bolt.Tx) error {

		b, err := tx.Bucket(users).CreateBucketIfNotExists([]byte(name))
		log.Check(log.WarnLevel, "Creating users subbucket: "+string(name), err)
		b.Put([]byte("key"), key)

		return nil
	})
}

func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(users).Bucket([]byte(name)); b != nil {
			if value := b.Get([]byte("key")); value != nil {
				key = string(value)
			}
		}
		return nil
	})
	return key
}

func SaveToken(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		if b, _ := tx.Bucket(tokens).CreateBucketIfNotExists([]byte(token)); b != nil {
			b.Put([]byte("name"), []byte(name))
			now, _ := time.Now().MarshalText()
			b.Put([]byte("date"), now)
		}
		return nil
	})
}

func CheckToken(token string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(tokens).Bucket([]byte(token)); b != nil {
			date := new(time.Time)
			date.UnmarshalText(b.Get([]byte("date")))
			if date.Add(time.Minute * 60).Before(time.Now()) {
				return nil
			}
			if value := b.Get([]byte("name")); value != nil {
				name = string(value)
			}
		}
		return nil
	})
	return name
}

func SaveAuthID(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(authid).Put([]byte(token), []byte(name))
		return nil
	})
}

func CheckAuthID(token string) (name string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(authid)
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
			b.Delete([]byte(token))
		}
		return nil
	})
	return name
}
