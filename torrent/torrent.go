// Package torrent provides function for operating BitTorrent client and torrent files.
// Client allows to seed and download files.
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
	// List of predefined torrent peers that can operate even without trackers. CDN nodes.
	list = []torrent.Peer{
		{IP: net.ParseIP("eu0.cdn.subut.ai"), Port: 50007},
		{IP: net.ParseIP("us0.cdn.subut.ai"), Port: 50007},
		{IP: net.ParseIP("us1.cdn.subut.ai"), Port: 50007},
		{IP: net.ParseIP("kg.cdn.subut.ai"), Port: 50007},
	}

	// List of torrent trackers that will be used in torrent files.
	builtinAnnounceList = [][]string{
		{"http://eu0.cdn.subut.ai:6882/announce"},
		{"http://us0.cdn.subut.ai:6882/announce"},
		{"http://us1.cdn.subut.ai:6882/announce"},
		{"http://kg.cdn.subut.ai:6882/announce"},
	}

	// Torrent client seeds and downloads files.
	client = initClient()
)

func initClient() *torrent.Client {
	err := os.MkdirAll(config.Storage.Path, 0600)
	log.Check(log.ErrorLevel, "Creating storage path", err)
	cl, err := torrent.NewClient(&torrent.Config{
		DataDir:           config.Storage.Path,
		Seed:              true,
		DisableEncryption: true,
		Debug:             false,
		NoDHT:             true,
	})

	log.Check(log.FatalLevel, "Creating torrent client", err)
	return cl
}

// Load returns torrent file for template. It tries to get torrent file from DB first.
// If no record found in DB it will generate new torrent file from template on disk.
// After generating new torrent file Load will store it in DB for future usage.
func Load(hash []byte) *bytes.Reader {
	file := db.Torrent(hash)
	if file == nil {
		var buf bytes.Buffer
		tfile := bufio.NewWriter(&buf)

		metaInfo := &metainfo.MetaInfo{AnnounceList: builtinAnnounceList}
		metaInfo.SetDefaults()

		err := metaInfo.Info.BuildFromFilePath(config.Storage.Path + string(hash))
		if log.Check(log.DebugLevel, "Creating torrent from local file", err) && len(config.CDN.Node) > 0 {
			httpclient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
			resp, err := httpclient.Get(config.CDN.Node + "/kurjun/rest/template/torrent?id=" + string(hash))
			if !log.Check(log.WarnLevel, "Getting file from CDN", err) && resp.StatusCode == http.StatusOK {

				_, err = io.Copy(tfile, resp.Body)
				log.Check(log.DebugLevel, "Reading CDN response to torrent file", err)

				err = resp.Body.Close()
				log.Check(log.DebugLevel, "Closing CDN response", err)
			} else {
				return nil
			}
		} else {
			err = metaInfo.Write(tfile)
			log.Check(log.DebugLevel, "Writing torrent file to buffer", err)
		}
		db.SaveTorrent(hash, buf.Bytes())
		file = db.Torrent(hash)
	}
	return bytes.NewReader(file)
}

// AddTorrent starting downloading or seeding template file. It adds torrent file to the torrent client.
// Second adding the same torrent will not add another instance, it will be processed only once.
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

		t, err := client.AddTorrent(metaInfo)
		if !log.Check(log.WarnLevel, "Adding torrent file to client", err) {
			t.AddPeers(list)

			go func() {
				<-t.GotInfo()
				t.DownloadAll()
			}()
			t.Seeding()
		}
	}
}

// SeedLocal getting list of all local template files and starts seeding it for other peers.
// It checks and adds new files every 60 seconds.
func SeedLocal() {
	for {
		for hash := range db.List() {
			if info := db.Info(hash); info["type"] == "template" {
				AddTorrent(hash)
			}
		}
		time.Sleep(time.Second * 60)
	}
}

// Info shows information about download progress for request template file.
// It return JSON with total and finished bytes of file.
// Info also drops broken torrents from client if any of them will be found.
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

// Close correctly finishing torrent client work.
func Close() {
	client.Close()
}

// IsDownloaded shows if particular template was downloaded or it still in progress.
func IsDownloaded(hash string) bool {
	for _, t := range client.Torrents() {
		if t.Name() == hash && t.Info().TotalLength() == t.BytesCompleted() {
			return true
		}
	}
	return false
}
