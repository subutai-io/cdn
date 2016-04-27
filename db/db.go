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
		for _, b := range []string{bucket, search, users, tokens, authid} {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			log.Check(log.FatalLevel, "Creating bucket: "+b, err)
		}
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction"+bucket, err)
	return db
}

func Write(owner, key, value string, options ...map[string]string) {
	if len(owner) == 0 {
		owner = "public"
	}
	err := db.Update(func(tx *bolt.Tx) error {
		now, _ := time.Now().MarshalText()

		b, err := tx.Bucket([]byte(users)).CreateBucketIfNotExists([]byte(owner))
		log.Check(log.WarnLevel, "Creating users subbucket: "+key, err)
		b, err = b.CreateBucketIfNotExists([]byte("files"))
		log.Check(log.WarnLevel, "Creating users:files subbucket: "+key, err)
		b.Put([]byte(key), []byte(value))

		b, err = tx.Bucket([]byte(bucket)).CreateBucketIfNotExists([]byte(key))
		log.Check(log.WarnLevel, "Creating subbucket: "+key, err)
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
		log.Check(log.WarnLevel, "Creating subbucket: "+key, err)
		b.Put(now, []byte(key))

		return nil
	})
	log.Check(log.WarnLevel, "Writing data to db", err)
}

func Delete(key string) (err error) {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)); b != nil {
			err = b.DeleteBucket([]byte(key))
		}
		return nil
	})
	return err
}

func Read(key string) (val string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)).Bucket([]byte(key)); b != nil {
			if value := b.Get([]byte("name")); value != nil {
				val = string(value)
			}
		}
		return nil
	})
	return val
}

func List() map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)); b != nil {
			b.ForEach(func(k, v []byte) error {
				if value := b.Bucket(k).Get([]byte("name")); value != nil {
					list[string(k)] = string(value)
				}
				return nil
			})
		}
		return nil
	})
	return list
}

func Info(hash string) map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(bucket)).Bucket([]byte(hash)); b != nil {
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
		if b := tx.Bucket([]byte(search)); b != nil {
			c := b.Cursor()
			for k, _ := c.Seek([]byte(query)); len(k) > 0 && bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
				b.Bucket(k).ForEach(func(kk, vv []byte) error {
					list[string(vv)] = string(k)
					return nil
				})
			}
		}
		return nil
	})
	return list
}

func LastHash(name string) (hash string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(search)).Bucket([]byte(name)); b != nil {
			_, v := b.Cursor().Last()
			hash = string(v)
		}
		return nil
	})
	return hash
}

func RegisterUser(name, key []byte) {
	db.Update(func(tx *bolt.Tx) error {

		b, err := tx.Bucket([]byte(users)).CreateBucketIfNotExists([]byte(name))
		log.Check(log.WarnLevel, "Creating users subbucket: "+string(name), err)
		b.Put([]byte("key"), key)

		return nil
	})
}

func UserKey(name string) (key string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(users)).Bucket([]byte(name)); b != nil {
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
		if b, _ := tx.Bucket([]byte(tokens)).CreateBucketIfNotExists([]byte(token)); b != nil {
			b.Put([]byte("name"), []byte(name))
			now, _ := time.Now().MarshalText()
			b.Put([]byte("date"), now)
		}
		return nil
	})
}

func CheckToken(token string) (name string) {
	db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(tokens)).Bucket([]byte(token)); b != nil {
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
		if b := tx.Bucket([]byte(authid)); b != nil {
			b.Put([]byte(token), []byte(name))
		}
		return nil
	})
}

func CheckAuthID(token string) (name string) {
	db.Update(func(tx *bolt.Tx) error {
		if b := tx.Bucket([]byte(authid)); b != nil {
			if value := b.Get([]byte(token)); value != nil {
				name = string(value)
				b.Delete([]byte(token))
			}
		}
		return nil
	})
	return name
}
