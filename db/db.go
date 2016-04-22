package db

import (
	"bytes"
	"time"

	"github.com/boltdb/bolt"

	"github.com/subutai-io/base/agent/log"
)

var (
	bucket = "MyBucket"
	search = "SearchIndex"
	users  = "Users"
	tokens = "Tokens"
	authid = "AuthID"
	db     = initdb()
)

func initdb() *bolt.DB {
	db, err := bolt.Open("my.db", 0600, nil)
	log.Check(log.FatalLevel, "Openning DB: my.db", err)
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		log.Check(log.FatalLevel, "Creating data bucket: "+bucket, err)

		_, err = tx.CreateBucketIfNotExists([]byte(search))
		log.Check(log.FatalLevel, "Creating search bucket: "+search, err)

		_, err = tx.CreateBucketIfNotExists([]byte(users))
		log.Check(log.FatalLevel, "Creating users bucket: "+search, err)

		_, err = tx.CreateBucketIfNotExists([]byte(tokens))
		log.Check(log.FatalLevel, "Creating users bucket: "+search, err)

		_, err = tx.CreateBucketIfNotExists([]byte(authid))
		log.Check(log.FatalLevel, "Creating users bucket: "+search, err)

		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction"+bucket, err)

	return db
}

func Write(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "subutai"
	}
	now, _ := time.Now().MarshalText()
	err := db.Update(func(tx *bolt.Tx) error {

		b, err := tx.Bucket([]byte(users)).CreateBucketIfNotExists([]byte(owner))
		log.Check(log.FatalLevel, "Creating users subbucket: "+key, err)
		b, err = b.CreateBucketIfNotExists([]byte("files"))
		log.Check(log.FatalLevel, "Creating users:files subbucket: "+key, err)
		b.Put([]byte(key), []byte(value))

		b, err = tx.Bucket([]byte(bucket)).CreateBucketIfNotExists([]byte(key))
		log.Check(log.FatalLevel, "Creating subbucket: "+key, err)
		b.Put([]byte("date"), now)
		b.Put([]byte("name"), []byte(value))
		b.Put([]byte("owner"), []byte(owner))

		for i, _ := range options {
			for k, v := range options[i] {
				err = b.Put([]byte(k), []byte(v))
				log.Check(log.WarnLevel, "Storing key: "+k, err)
			}
		}

		b, err = tx.Bucket([]byte(search)).CreateBucketIfNotExists([]byte(value))
		log.Check(log.FatalLevel, "Creating subbucket: "+key, err)
		b.Put(now, []byte(key))

		return nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Read(key string) (val string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket)).Bucket([]byte(key))
		if value := b.Get([]byte("name")); value != nil {
			val = string(value)
		}
		return nil
	})
	return val
}

func List() map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
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

func Close() {
	db.Close()
}

func Search(query string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(search))
		c := b.Cursor()
		for k, _ := c.Seek([]byte(query)); bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
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
		b := tx.Bucket([]byte(search)).Bucket([]byte(name))
		_, v := b.Cursor().Last()
		hash = string(v)
		return nil
	})
	return hash
}

func RegisterUser(name, key []byte) {
	db.Update(func(tx *bolt.Tx) error {

		b, err := tx.Bucket([]byte(users)).CreateBucketIfNotExists([]byte(name))
		log.Check(log.FatalLevel, "Creating users subbucket: "+string(name), err)
		b.Put([]byte("key"), key)

		return nil
	})
}

func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(users)).Bucket([]byte(name))
		if value := b.Get([]byte("key")); value != nil {
			key = string(value)
		}
		return nil
	})
	return key
}

func SaveToken(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tokens))
		b.Put([]byte(token), []byte(name))
		return nil
	})
}

func CheckToken(token string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(tokens))
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
		}
		return nil
	})
	return name
}

func SaveAuthID(name, token string) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(authid))
		b.Put([]byte(token), []byte(name))
		return nil
	})
}

func CheckAuthID(token string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(authid))
		if value := b.Get([]byte(token)); value != nil {
			name = string(value)
		}
		return nil
	})
	return name
}
