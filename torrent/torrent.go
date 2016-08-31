package torrent

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/subutai-io/base/agent/log"
	"github.com/subutai-io/gorjun/config"
	"github.com/subutai-io/gorjun/db"
)

var (
	list = []torrent.Peer{
		{IP: net.ParseIP("52.28.78.136"), Port: 50007},    // eu0.cdn.subut.ai
		{IP: net.ParseIP("52.90.197.198"), Port: 50007},   // us0.cdn.subut.ai
		{IP: net.ParseIP("54.183.100.182"), Port: 50007},  // us1.cdn.subut.ai
		{IP: net.ParseIP("158.181.187.116"), Port: 50007}, // kg.cdn.subut.ai
	}

	builtinAnnounceList = [][]string{
		{"http://eu0.cdn.subut.ai:6882/announce"},
		{"http://us0.cdn.subut.ai:6882/announce"},
		{"http://us1.cdn.subut.ai:6882/announce"},
		{"http://kg.cdn.subut.ai:6882/announce"},
	}

	client = initClient()
)

func initClient() *torrent.Client {
	os.MkdirAll(config.Storage.Path+"/p2p/", 0600)
	client, err := torrent.NewClient(&torrent.Config{
		DataDir:           config.Storage.Path,
		Seed:              true,
		DisableEncryption: true,
		Debug:             false,
		NoDHT:             true,
	})

	log.Check(log.FatalLevel, "Creating torrent client", err)
	return client
}

func Load(hash []byte) *bytes.Reader {
	file := db.Torrent(hash)
	if file == nil {
		var buf bytes.Buffer
		tfile := bufio.NewWriter(&buf)

		metaInfo := &metainfo.MetaInfo{AnnounceList: builtinAnnounceList}
		metaInfo.SetDefaults()

		err := metaInfo.Info.BuildFromFilePath(config.Storage.Path + string(hash))
		if log.Check(log.DebugLevel, "Creating torrent from local file", err) {
			httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			resp, err := httpclient.Get(config.CDN.Node + "/kurjun/rest/template/torrent?id=" + string(hash))
			if !log.Check(log.WarnLevel, "Getting file from CDN", err) && resp.StatusCode == http.StatusOK {
				io.Copy(tfile, resp.Body)
				resp.Body.Close()
			} else {
				return nil
			}
		} else {
			metaInfo.Write(tfile)
		}
		db.SaveTorrent(hash, buf.Bytes())
		file = db.Torrent(hash)
	}
	return bytes.NewReader(file)
}

func AddTorrent(hash string) {
	reader := Load([]byte(hash))
	if reader == nil {
		return
	}
	metaInfo, err := metainfo.Load(reader)
	if log.Check(log.InfoLevel, "Creating torrent for "+hash, err) {
		return
	}
	_, ok := client.Torrent(metaInfo.Info.Hash())
	if !ok {
		metaInfo.AnnounceList = builtinAnnounceList
		metaInfo.SetDefaults()
		metaInfo.Info.MarshalBencode()

		t, _ := client.AddTorrent(metaInfo)
		t.AddPeers(list)

		go func() {
			<-t.GotInfo()
			t.DownloadAll()
		}()
		t.Seeding()
	}
}

func SeedLocal() {
	for {
		for hash, _ := range db.List() {
			if info := db.Info(hash); info["type"] == "template" {
				AddTorrent(hash)
			}
		}
		time.Sleep(time.Second * 60)
	}
}

func Info(id string) (output string) {
	for _, t := range client.Torrents() {
		if t.Info().TotalLength() != 0 {
			if t.Name() == id {
				output = fmt.Sprintf(`{"total":%d,"done":%d}`, t.Info().TotalLength(), t.BytesCompleted())
			}
		} else {
			t.Drop()
		}
	}
	return output
}

func Close() {
	client.Close()
}

func IsDownloaded(hash string) bool {
	for _, t := range client.Torrents() {
		if t.Name() == hash && t.Info().TotalLength() == t.BytesCompleted() {
			return true
		}
	}
	return false
}
