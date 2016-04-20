package db

import (
	"bytes"

	"github.com/boltdb/bolt"

	"github.com/subutai-io/base/agent/log"
)

var (
	bucket = "MyBucket"
	search = "SearchIndex"
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

		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction"+bucket, err)

	return db
}

func Write(key, value string, options ...map[string]string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket([]byte(bucket)).CreateBucketIfNotExists([]byte(key))
		log.Check(log.FatalLevel, "Creating subbucket: "+key, err)
		err = b.Put([]byte("name"), []byte(value))
		log.Check(log.WarnLevel, "Storing key: "+key, err)

		for i, _ := range options {
			for k, v := range options[i] {
				err = b.Put([]byte(k), []byte(v))
				log.Check(log.WarnLevel, "Storing key: "+k, err)
				log.Info(k + ": " + v)
			}
		}

		b, err = tx.Bucket([]byte(search)).CreateBucketIfNotExists([]byte(value))
		log.Check(log.FatalLevel, "Creating subbucket: "+key, err)
		err = b.Put([]byte(key), []byte(""))
		log.Check(log.WarnLevel, "Storing key: "+key, err)

		return nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Read(key string) (value string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket)).Bucket([]byte(key))
		value = string(b.Get([]byte("name")))
		return nil
	})
	return value
}

func List() map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.ForEach(func(k, v []byte) error {
			list[string(k)] = string(b.Bucket(k).Get([]byte("name")))
			b.Bucket(k).ForEach(func(kk, vv []byte) error {
				log.Info(string(kk) + ": " + string(vv))
				return nil
			})
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

		c := tx.Bucket([]byte(search)).Cursor()
		for k, _ := c.Seek([]byte(query)); bytes.HasPrefix(k, []byte(query)); k, _ = c.Next() {
			b.Bucket(k).ForEach(func(kk, vv []byte) error {
				list[string(kk)] = string(k)
				return nil
			})
		}
		return nil
	})
	return list
}
