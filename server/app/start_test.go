package app

import (
	"testing"
	"runtime"
	"github.com/subutai-io/agent/log"
	"fmt"
)

func TestRunServer(t *testing.T) {
	Integration = 0
	SetUp()
	defer TearDown()
	tests := []struct {
		name string
	}{
		{"TestRunServer-"},
		// TODO: Add test cases.
	}
	tests[0].name += "0"
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
