package app

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"os"

	"github.com/boltdb/bolt"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/db"
)

var (
	verifiedUsers = []string{"subutai", "jenkins", "docker", "travis", "appveyor", "devops"}
)

func CheckOwner(owner string) bool {
	exists := false
	db.DB.View(func(tx *bolt.Tx) error {
		exists = tx.Bucket(db.Users).Bucket([]byte(owner)) != nil
		return nil
	})
	return exists
}

func CheckToken(token string) bool {
	return db.TokenOwner(token) != ""
}

func Hash(file string, algo string) string {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to open file %s to calculate hash", file))
		return ""
	}
	hash := md5.New()
	switch algo {
	case "md5":
		hash = md5.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	}
	if _, err := io.Copy(hash, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func In(item string, list []string) bool {
	for _, v := range list {
		if item == v {
			return true
		}
	}
	return false
}
