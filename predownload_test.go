package app

import (
	"fmt"
	"os/exec"
	"github.com/subutai-io/agent/log"
	"github.com/subutai-io/cdn/libgorjun"
)

func PreDownloadFiles(scope int, user gorjun.GorjunUser) { // private either 0 (false) or 1 (true)
	for i := 0; i < len(Files[scope][user.Username][IDsLayer]); i++ {
		cmd := exec.Command("wget", "-O", Dirs[scope][user.Username] + Files[scope][user.Username][NamesLayer][i], Raw + Files[scope][user.Username][IDsLayer][i])
		cmd.Run()
	}
	log.Info(fmt.Sprintf("All %s files of user %s downloaded", FileType(scope), user.Username))
	return
}

func PreDownloadAllFiles(user gorjun.GorjunUser) {
	PreDownloadFiles(PublicScope, user)
	PreDownloadFiles(PrivateScope, user)
	log.Info(fmt.Sprintf("All files of user %s downloaded", user.Username))
	return
}

func PreDownload() {
	log.Info("Pre-downloading files to CDN")
	PreDownloadAllFiles(Subutai)
	PreDownloadAllFiles(Lorem)
	PreDownloadAllFiles(Ipsum)
	log.Info("Pre-downloading files finished")
}
