package app

import (
	"testing"
	"runtime"
	"github.com/subutai-io/agent/log"
	"fmt"
)

func TestRunServer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"TestRunServer-1"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RunServer()
			log.Info("Server has started successfully")
			if runtime.NumGoroutine() == 0 {
				log.Warn(fmt.Sprintf("Number of goroutines: %d", runtime.NumGoroutine()))
				t.Errorf("Number of goroutines isn't two")
			}
			for {
				if Stop != nil {
					log.Info("Stop request sent")
					Stop <- true
					break
				}
			}
			<-Stop
			close(Stop)
		})
	}
}
