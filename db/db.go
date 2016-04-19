package db

import (
	"github.com/boltdb/bolt"

	"github.com/subutai-io/base/agent/log"
)

var (
	bucket = "MyBucket"
	db     = initdb()
)

func initdb() *bolt.DB {
	db, err := bolt.Open("my.db", 0600, nil)
	log.Check(log.FatalLevel, "Openning DB: my.db", err)
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		log.Check(log.FatalLevel, "Creating bucket: "+bucket, err)
		return nil
	})
	log.Check(log.FatalLevel, "Finishing update transaction"+bucket, err)

	return db
}

func Write(key, value string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		err := b.Put([]byte(key), []byte(value))
		log.Check(log.WarnLevel, "Creating bucket: "+bucket, err)
		return nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}
}

func Read(key string) (value string) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		value = string(b.Get([]byte(key)))
		return nil
	})
	return value
}

func List() map[string]string {
	list := make(map[string]string)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.ForEach(func(k, v []byte) error {
			list[string(k)] = string(v)
			return nil
		})
		return nil
	})
	return list
}

func Close() {
	db.Close()
}
