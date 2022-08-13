package discord

import (
	"sync/atomic"
)

var (
	DeadTokens    int32
	SuccessDM     int32
	FailedDM      int32
	TotalRequests int32

	//24hours
	PsuccessDM     int32
	PfailedDM      int32
	PtotalRequests int32
)

func AddSuccessDM() {
	atomic.AddInt32(&SuccessDM, 1)
	atomic.AddInt32(&PsuccessDM, 1)
}
func AddFailedDM() {
	atomic.AddInt32(&FailedDM, 1)
	atomic.AddInt32(&PfailedDM, 1)
}
func AddTotalRequests() {
	atomic.AddInt32(&TotalRequests, 1)
	atomic.AddInt32(&PtotalRequests, 1)
}
