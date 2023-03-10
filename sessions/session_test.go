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

package sessions

import (
	"reflect"
	"testing"
	"time"

	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func TestSessionIDCGRID(t *testing.T) {
	//empty check
	sessionID := new(SessionID)
	rcv := sessionID.CGRID()
	eOut := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	sessionID.OriginHost = "testhost"
	sessionID.OriginID = "testid"
	rcv = sessionID.CGRID()
	eOut = "2aaff7e3e832de08b0604a79a18ccc6bba823360"
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestSessionCgrID(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.cgrID()
	eOut := ""
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	session.CGRID = "testID"
	eOut = "testID"
	rcv = session.cgrID()
	if !reflect.DeepEqual(eOut, rcv) && session.CGRID == "testID" {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestSessionClone(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.Clone()
	eOut := new(Session)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	tTime := time.Now()
	tTime2 := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
	session = &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: time.Duration(18),
		SRuns: []*SRun{
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Duration(1),
				LastUsage:     time.Duration(2),
				TotalUsage:    time.Duration(3),
				NextAutoDebit: &tTime,
			},
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test2"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID2"},
				ExtraDuration: time.Duration(4),
				LastUsage:     time.Duration(5),
				TotalUsage:    time.Duration(6),
				NextAutoDebit: &tTime2,
			},
		},
	}
	eOut = &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: time.Duration(18),
		SRuns: []*SRun{
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Duration(1),
				LastUsage:     time.Duration(2),
				TotalUsage:    time.Duration(3),
				NextAutoDebit: &tTime,
			},
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test2"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID2"},
				ExtraDuration: time.Duration(4),
				LastUsage:     time.Duration(5),
				TotalUsage:    time.Duration(6),
				NextAutoDebit: &tTime2,
			},
		},
	}
	rcv = session.Clone()
	if !reflect.DeepEqual(eOut, rcv) && session.CGRID == "testID" {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//check clone
	rcv.CGRID = "newCGRID"

	if session.CGRID == "newCGRID" {
		t.Errorf("Expecting: CGRID, received: newCGRID")
	}
	rcv.SRuns[1].TotalUsage = time.Duration(10)
	if session.SRuns[1].TotalUsage == time.Duration(10) {
		t.Errorf("Expecting: %s, received: %s", time.Duration(3), time.Duration(10))
	}
	tTimeNow := time.Now()
	*rcv.SRuns[1].NextAutoDebit = tTimeNow
	if *session.SRuns[1].NextAutoDebit == tTimeNow {
		t.Errorf("Expecting: %s, received: %s", time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC), tTimeNow)

	}

}

//Test1 ExtraDuration 0 and LastUsage < initial

//Test1 ExtraDuration 0 and LastUsage < initial
func TestSRunDebitReserve(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(0),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	//start with extraDuration 0 and the difference go in rDur
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test2 ExtraDuration 0 and LastUsage > initial
func TestSRunDebitReserve2(t *testing.T) {
	lastUsage := time.Duration(2*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(0),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test3 ExtraDuration ( 1m < duration) and LastUsage < initial
func TestSRunDebitReserve3(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != (duration - lastUsage) {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test4 ExtraDuration 1m and LastUsage > initial
func TestSRunDebitReserve4(t *testing.T) {
	lastUsage := time.Duration(2*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	//We have extraDuration 1 minute and 30s different
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != time.Duration(1*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(1*time.Minute+30*time.Second), rDur)
	}
	if sr.ExtraDuration != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), sr.ExtraDuration)
	}
	if sr.LastUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.LastUsage)
	}
	if sr.TotalUsage != lastUsage {
		t.Errorf("Expecting: %+v, received: %+v", lastUsage, sr.TotalUsage)
	}
}

//Test5 ExtraDuration 3m ( > initialDuration) and LastUsage < initial
func TestSRunDebitReserve5(t *testing.T) {
	lastUsage := time.Duration(1*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(3 * time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	//in debit reserve we start with an extraDuration 3m
	//after we add the different dur-lastUsed (+30s)
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), rDur)
	}
	//ExtraDuration (3m30s - 2m)
	if sr.ExtraDuration != time.Duration(1*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(1*time.Minute+30*time.Second), sr.ExtraDuration)
	}
	if sr.LastUsage != duration {
		t.Errorf("Expecting: %+v, received: %+v", duration, sr.LastUsage)
	}
	if sr.TotalUsage != time.Duration(3*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(3*time.Minute+30*time.Second), sr.TotalUsage)
	}
}

//Test6 ExtraDuration 3m ( > initialDuration) and LastUsage > initial
func TestSRunDebitReserve6(t *testing.T) {
	lastUsage := time.Duration(2*time.Minute + 30*time.Second)
	duration := time.Duration(2 * time.Minute)
	sr := &SRun{
		ExtraDuration: time.Duration(3 * time.Minute),
		LastUsage:     duration,
		TotalUsage:    duration,
	}
	//in debit reserve we start with an extraDuration 3m
	//after we add the different dur-lastUsed (-30s)
	if rDur := sr.debitReserve(duration, &lastUsage); rDur != time.Duration(0) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(0), rDur)
	}
	//ExtraDuration (2m30s - 2m)
	if sr.ExtraDuration != time.Duration(30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(30*time.Second), sr.ExtraDuration)
	}
	if sr.LastUsage != duration {
		t.Errorf("Expecting: %+v, received: %+v", duration, sr.LastUsage)
	}
	// 2m(initial Total) + 2m30s(correction)
	if sr.TotalUsage != time.Duration(4*time.Minute+30*time.Second) {
		t.Errorf("Expecting: %+v, received: %+v", time.Duration(4*time.Minute+30*time.Second), sr.TotalUsage)
	}
}

func TestSessionAsCGREventsRawEvent(t *testing.T) {
	ev := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.12,
	}
	s := &Session{
		CGRID:      "RandomCGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(ev),
	}
	if cgrEvs, _ := s.asCGREvents(); len(cgrEvs) != 0 {
		t.Errorf("Expecting: 1, received: %+v", len(cgrEvs))
	}

}

func TestSessionAsCGREvents(t *testing.T) {
	startEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.12,
	}
	ev := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.RunID:       utils.MetaDefault,
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.13,
	}
	s := &Session{
		CGRID:      "RandomCGRID",
		Tenant:     "cgrates.org",
		EventStart: engine.NewMapEvent(startEv),
		SRuns: []*SRun{
			&SRun{
				Event:      engine.NewMapEvent(ev),
				TotalUsage: time.Duration(2 * time.Second),
			},
		},
	}
	//check for some fields if populated correct
	cgrEvs, err := s.asCGREvents()
	if err != nil {
		t.Error(err)
	} else if len(cgrEvs) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(cgrEvs))
	}
	if cgrEvs[0].Event[utils.RunID] != utils.MetaDefault {
		t.Errorf("Expecting: %+v, received: %+v", utils.MetaDefault, cgrEvs[1].Event[utils.RunID])
	} else if cgrEvs[0].Event[utils.Cost] != 12.13 {
		t.Errorf("Expecting: %+v, received: %+v", 12.13, cgrEvs[1].Event[utils.Cost])
	}
}

func TestSessionAsExternalSessions(t *testing.T) {
	startEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.12,
	}
	ev := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.RunID:       utils.MetaDefault,
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.13,
	}
	tTime := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
	s := &Session{
		CGRID:         "RandomCGRID",
		Tenant:        "cgrates.org",
		EventStart:    engine.NewMapEvent(startEv),
		DebitInterval: time.Second,
		SRuns: []*SRun{
			&SRun{
				Event:         engine.NewMapEvent(ev),
				TotalUsage:    time.Duration(2 * time.Second),
				NextAutoDebit: &tTime,
			},
		},
	}
	exp := []*ExternalSession{
		&ExternalSession{
			CGRID:    "RandomCGRID",
			RunID:    utils.MetaDefault,
			ToR:      utils.VOICE,
			OriginID: "123451",
			// OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:      utils.SessionS + "_" + "TEST_EVENT",
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1004",
			SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			Usage:       time.Duration(2 * time.Second),
			ExtraFields: map[string]string{
				utils.EVENT_NAME: "TEST_EVENT2",
			},
			NodeID:        "ALL",
			DebitInterval: time.Second,
			NextAutoDebit: tTime,
			// aSs[i].LoopIndex:     sr.CD.LoopIndex,
			// aSs[i].DurationIndex: sr.CD.DurationIndex,
			// aSs[i].MaxRate:       sr.CD.MaxRate,
			// aSs[i].MaxRateUnit:   sr.CD.MaxRateUnit,
			// aSs[i].MaxCostSoFar:  sr.CD.MaxCostSoFar,
		},
	}
	//check for some fields if populated correct
	rply := s.AsExternalSessions("", "ALL")
	if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

}

func TestSessionAsExternalSessions2(t *testing.T) {
	startEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.12,
	}
	ev := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.RunID:       utils.MetaDefault,
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.13,
	}
	s := &Session{
		CGRID:         "RandomCGRID",
		Tenant:        "cgrates.org",
		EventStart:    engine.NewMapEvent(startEv),
		DebitInterval: time.Second,
		SRuns: []*SRun{
			&SRun{
				Event:      engine.NewMapEvent(ev),
				TotalUsage: time.Duration(2 * time.Second),
				CD: &engine.CallDescriptor{
					LoopIndex:     10,
					DurationIndex: 3 * time.Second,
					MaxRate:       11,
					MaxRateUnit:   30 * time.Second,
					MaxCostSoFar:  20,
				},
			},
		},
	}
	exp := []*ExternalSession{
		&ExternalSession{
			CGRID:    "RandomCGRID",
			RunID:    utils.MetaDefault,
			ToR:      utils.VOICE,
			OriginID: "123451",
			// OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
			Source:      utils.SessionS + "_" + "TEST_EVENT",
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1004",
			SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
			AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			Usage:       time.Duration(2 * time.Second),
			ExtraFields: map[string]string{
				utils.EVENT_NAME: "TEST_EVENT2",
			},
			NodeID:        "ALL",
			DebitInterval: time.Second,
			LoopIndex:     10,
			DurationIndex: 3 * time.Second,
			MaxRate:       11,
			MaxRateUnit:   30 * time.Second,
			MaxCostSoFar:  20,
		},
	}
	//check for some fields if populated correct
	rply := s.AsExternalSessions("", "ALL")
	if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

}

func TestSessionAsExternalSessions3(t *testing.T) {
	startEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.12,
	}
	ev := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT2",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.RunID:       utils.MetaDefault,
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:       time.Duration(2 * time.Second),
		utils.Cost:        12.13,
	}
	tTime := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)

	s := &Session{
		CGRID:         "RandomCGRID",
		Tenant:        "cgrates.org",
		EventStart:    engine.NewMapEvent(startEv),
		DebitInterval: time.Second,
		SRuns: []*SRun{
			&SRun{
				Event:      engine.NewMapEvent(ev),
				TotalUsage: time.Duration(2 * time.Second),
				CD: &engine.CallDescriptor{
					LoopIndex:     10,
					DurationIndex: 3 * time.Second,
					MaxRate:       11,
					MaxRateUnit:   30 * time.Second,
					MaxCostSoFar:  20,
				},
				NextAutoDebit: &tTime,
			},
		},
	}
	exp := &ExternalSession{
		CGRID:    "RandomCGRID",
		RunID:    utils.MetaDefault,
		ToR:      utils.VOICE,
		OriginID: "123451",
		// OriginHost:  s.EventStart.GetStringIgnoreErrors(utils.OriginHost),
		Source:      utils.SessionS + "_" + "TEST_EVENT",
		RequestType: utils.META_PREPAID,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1004",
		SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
		AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		Usage:       time.Duration(2 * time.Second),
		ExtraFields: map[string]string{
			utils.EVENT_NAME: "TEST_EVENT2",
		},
		NodeID:        "ALL",
		DebitInterval: time.Second,
		LoopIndex:     10,
		DurationIndex: 3 * time.Second,
		MaxRate:       11,
		MaxRateUnit:   30 * time.Second,
		MaxCostSoFar:  20,
		NextAutoDebit: tTime,
	}
	//check for some fields if populated correct
	rply := s.AsExternalSession(s.SRuns[0], "", "ALL")
	if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

}

func TestSessiontotalUsage(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.totalUsage()
	eOut := time.Duration(0)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	tTime := time.Now()
	tTime2 := time.Date(2020, time.April, 18, 23, 0, 0, 0, time.UTC)
	session = &Session{
		CGRID:         "CGRID",
		Tenant:        "cgrates.org",
		ResourceID:    "resourceID",
		ClientConnID:  "ClientConnID",
		EventStart:    engine.NewMapEvent(nil),
		DebitInterval: time.Duration(18),
		SRuns: []*SRun{
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID"},
				ExtraDuration: time.Duration(1),
				LastUsage:     time.Duration(2),
				TotalUsage:    time.Duration(5),
				NextAutoDebit: &tTime,
			},
			{Event: engine.NewMapEvent(nil),
				CD:            &engine.CallDescriptor{Category: "test2"},
				EventCost:     &engine.EventCost{CGRID: "testCGRID2"},
				ExtraDuration: time.Duration(4),
				LastUsage:     time.Duration(5),
				TotalUsage:    time.Duration(6),
				NextAutoDebit: &tTime2,
			},
		},
	}
	eOut = time.Duration(5)
	rcv = session.totalUsage()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestSessionstopSTerminator(t *testing.T) {
	//empty check
	session := new(Session)
	rcv := session.totalUsage()
	eOut := time.Duration(0)
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//normal check
	session = &Session{
		sTerminator: &sTerminator{endChan: make(chan struct{})},
	}
	session.stopSTerminator()
	if session.sTerminator.endChan != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(session.sTerminator.endChan))
	}
}

func TestSessionstopDebitLoops(t *testing.T) {
	session := &Session{
		debitStop: make(chan struct{}),
	}
	session.stopDebitLoops()
	if session.debitStop != nil {
		t.Errorf("Expecting: nil, received: %s", utils.ToJSON(session.debitStop))
	}

}
