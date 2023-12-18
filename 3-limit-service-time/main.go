//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"fmt"
	"time"
)

const maxFreeProcessingTimeSeconds = 10

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	ticker := time.NewTicker(time.Second)
	doneCh := make(chan struct{})
	defer func() {
		// close(doneCh)
		ticker.Stop()
		// recover()
	}()

	start := time.Now()

	go func() {
		for range ticker.C {
			u.TimeUsed += 1
			if !u.IsPremium && u.TimeUsed >= maxFreeProcessingTimeSeconds-1 {
				fmt.Println("Free processing time is over for UserID:", u.ID)

				doneCh <- struct{}{}
				return
			}
		}
	}()

	go func() {
		process()
		doneCh <- struct{}{}
	}()

	<-doneCh

	timeSpentSeconds := int64(time.Since(start).Seconds())
	fmt.Printf("Processed %ds (total %ds) for UserID: %d\n", timeSpentSeconds, u.TimeUsed, u.ID)

	return true
}

func main() {
	RunMockServer()
}
