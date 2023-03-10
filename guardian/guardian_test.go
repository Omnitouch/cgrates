/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package guardian

import (
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Omnitouch/cgrates/utils"
)

func delayHandler() (interface{}, error) {
	time.Sleep(100 * time.Millisecond)
	return nil, nil
}

// Forks 3 groups of workers and makes sure that the time for execution is the one we expect for all 15 goroutines (with 100ms )
func TestGuardianMultipleKeys(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian.Guard(delayHandler, 0, key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*100) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur ||
		execTime > mustExecDur+time.Duration(100*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian.lkMux.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian.locks[key]; hasKey {
			t.Errorf("Possible memleak for key: %s", key)
		}
	}
	Guardian.lkMux.Unlock()
}

func TestGuardianTimeout(t *testing.T) {
	tStart := time.Now()
	maxIter := 5
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	for i := 0; i < maxIter; i++ {
		for _, key := range keys {
			sg.Add(1)
			go func(key string) {
				Guardian.Guard(delayHandler, time.Duration(10*time.Millisecond), key)
				sg.Done()
			}(key)
		}
	}
	sg.Wait()
	mustExecDur := time.Duration(maxIter*10) * time.Millisecond
	if execTime := time.Now().Sub(tStart); execTime < mustExecDur ||
		execTime > mustExecDur+time.Duration(100*time.Millisecond) {
		t.Errorf("Execution took: %v", execTime)
	}
	Guardian.lkMux.Lock()
	for _, key := range keys {
		if _, hasKey := Guardian.locks[key]; hasKey {
			t.Error("Possible memleak")
		}
	}
	Guardian.lkMux.Unlock()
}

func TestGuardianGuardIDs(t *testing.T) {

	//lock with 3 keys
	lockIDs := []string{"test1", "test2", "test3"}
	// make sure the keys are not in guardian before lock
	Guardian.lkMux.Lock()
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locks[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	Guardian.lkMux.Unlock()
	// lock 3 items
	tStart := time.Now()
	lockDur := 2 * time.Millisecond
	Guardian.GuardIDs("", lockDur, lockIDs...)
	Guardian.lkMux.Lock()
	for _, lockID := range lockIDs {
		if itmLock, hasKey := Guardian.locks[lockID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lockID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected itmLock found: %+v", itmLock)
		}
	}
	Guardian.lkMux.Unlock()
	secLockDur := time.Duration(1 * time.Millisecond)
	// second lock to test counter
	go Guardian.GuardIDs("", secLockDur, lockIDs[1:]...)
	time.Sleep(30 * time.Microsecond) // give time for goroutine to lock
	// check if counters were properly increased
	Guardian.lkMux.Lock()
	lkID := lockIDs[0]
	eCnt := int64(1)
	if itmLock, hasKey := Guardian.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[1]
	eCnt = int64(2)
	if itmLock, hasKey := Guardian.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	lkID = lockIDs[2]
	eCnt = int64(1) // we did not manage to increase it yet since it did not pass first lock
	if itmLock, hasKey := Guardian.locks[lkID]; !hasKey {
		t.Errorf("Cannot find lock for lockID: %s", lkID)
	} else if itmLock.cnt != eCnt {
		t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
	}
	Guardian.lkMux.Unlock()
	time.Sleep(lockDur + secLockDur + 50*time.Millisecond) // give time to unlock before proceeding

	// make sure all counters were removed
	for _, lockID := range lockIDs {
		if _, hasKey := Guardian.locks[lockID]; hasKey {
			t.Errorf("Unexpected lockID found: %s", lockID)
		}
	}
	// test lock  without timer
	refID := Guardian.GuardIDs("", 0, lockIDs...)

	if totalLockDur := time.Now().Sub(tStart); totalLockDur < lockDur {
		t.Errorf("Lock duration too small")
	}
	time.Sleep(time.Duration(30) * time.Millisecond)
	// making sure the items stay locked
	Guardian.lkMux.Lock()
	if len(Guardian.locks) != 3 {
		t.Errorf("locks should have 3 elements, have: %+v", Guardian.locks)
	}
	for _, lkID := range lockIDs {
		if itmLock, hasKey := Guardian.locks[lkID]; !hasKey {
			t.Errorf("Cannot find lock for lockID: %s", lkID)
		} else if itmLock.cnt != 1 {
			t.Errorf("Unexpected counter: %d for itmLock with id %s", itmLock.cnt, lkID)
		}
	}
	Guardian.lkMux.Unlock()
	Guardian.UnguardIDs(refID)
	// make sure items were unlocked
	Guardian.lkMux.Lock()
	if len(Guardian.locks) != 0 {
		t.Errorf("locks should have 0 elements, has: %+v", Guardian.locks)
	}
	Guardian.lkMux.Unlock()
}

// TestGuardianGuardIDsConcurrent executes GuardIDs concurrently
func TestGuardianGuardIDsConcurrent(t *testing.T) {
	maxIter := 500
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	refID := utils.GenUUID()
	for i := 0; i < maxIter; i++ {
		sg.Add(1)
		go func() {
			if retRefID := Guardian.GuardIDs(refID, 0, keys...); retRefID != "" {
				if lkIDs := Guardian.UnguardIDs(refID); !reflect.DeepEqual(keys, lkIDs) {
					t.Errorf("expecting: %+v, received: %+v", keys, lkIDs)
				}
			}
			sg.Done()
		}()
	}
	sg.Wait()

	Guardian.lkMux.Lock()
	if len(Guardian.locks) != 0 {
		t.Errorf("Possible memleak for locks: %+v", Guardian.locks)
	}
	Guardian.lkMux.Unlock()
	Guardian.refsMux.Lock()
	if len(Guardian.refs) != 0 {
		t.Errorf("Possible memleak for refs: %+v", Guardian.refs)
	}
	Guardian.refsMux.Unlock()
}

func TestGuardianGuardIDsTimeoutConcurrent(t *testing.T) {
	maxIter := 50
	sg := new(sync.WaitGroup)
	keys := []string{"test1", "test2", "test3"}
	refID := utils.GenUUID()
	for i := 0; i < maxIter; i++ {
		sg.Add(1)
		go func() {
			Guardian.GuardIDs(refID, time.Duration(time.Microsecond), keys...)
			sg.Done()
		}()
	}
	sg.Wait()
	time.Sleep(10 * time.Millisecond)
	Guardian.lkMux.Lock()
	if len(Guardian.locks) != 0 {
		t.Errorf("Possible memleak for locks: %+v", Guardian.locks)
	}
	Guardian.lkMux.Unlock()
	Guardian.refsMux.Lock()
	if len(Guardian.refs) != 0 {
		t.Errorf("Possible memleak for refs: %+v", Guardian.refs)
	}
	Guardian.refsMux.Unlock()
}

// BenchmarkGuard-8      	  200000	     13759 ns/op
func BenchmarkGuard(b *testing.B) {
	for n := 0; n < b.N; n++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(time.Microsecond)
			return 0, nil
		}, 0, "1")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(time.Microsecond)
			return 0, nil
		}, 0, "2")
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(time.Microsecond)
			return 0, nil
		}, 0, "1")
	}

}

// BenchmarkGuardian-8   	 1000000	      5794 ns/op
func BenchmarkGuardian(b *testing.B) {
	for n := 0; n < b.N; n++ {
		go Guardian.Guard(func() (interface{}, error) {
			time.Sleep(time.Microsecond)
			return 0, nil
		}, 0, strconv.Itoa(n))
	}
}

// BenchmarkGuardIDs-8   	 1000000	      8732 ns/op
func BenchmarkGuardIDs(b *testing.B) {
	for n := 0; n < b.N; n++ {
		go func() {
			if refID := Guardian.GuardIDs("", 0, strconv.Itoa(n)); refID != "" {
				time.Sleep(time.Microsecond)
				Guardian.UnguardIDs(refID)
			}
		}()
	}
}
