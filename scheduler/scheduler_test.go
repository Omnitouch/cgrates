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
package scheduler

import (
	"testing"
	"time"

	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func TestSchedulerUpdateActStats(t *testing.T) {
	sched := &Scheduler{actStatsInterval: 10 * time.Millisecond, actSuccessStats: make(map[string]map[time.Time]bool)}
	sched.updateActStats(&engine.Action{Id: "REMOVE_1", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.REMOVE_ACCOUNT])
	}
	sched.updateActStats(&engine.Action{Id: "REMOVE_2", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats[utils.REMOVE_ACCOUNT])
	}
	sched.updateActStats(&engine.Action{Id: "LOG1", ActionType: utils.LOG}, false)
	if len(sched.actSuccessStats[utils.LOG]) != 1 ||
		len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 2 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
	time.Sleep(sched.actStatsInterval)
	sched.updateActStats(&engine.Action{Id: "REMOVE_3", ActionType: utils.REMOVE_ACCOUNT}, false)
	if len(sched.actSuccessStats[utils.REMOVE_ACCOUNT]) != 1 || len(sched.actSuccessStats) != 1 {
		t.Errorf("Wrong stats: %+v", sched.actSuccessStats)
	}
}
