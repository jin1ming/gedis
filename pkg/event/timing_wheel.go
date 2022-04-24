package event

import (
	"github.com/RussellLuo/timingwheel"
	"time"
)

var globalTimingWheel *timingwheel.TimingWheel

func init() {
	globalTimingWheel = timingwheel.NewTimingWheel(1*time.Second, 60)
	globalTimingWheel.Start()
}

func GetGlobalTimingWheel() *timingwheel.TimingWheel {
	return globalTimingWheel
}
