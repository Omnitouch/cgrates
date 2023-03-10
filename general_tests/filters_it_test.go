// +build integration

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

package general_tests

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	v1 "github.com/Omnitouch/cgrates/apier/v1"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

var (
	fltrCfgPath string
	fltrCfg     *config.CGRConfig
	fltrRpc     *rpc.Client
	fltrConfDIR string //run tests for specific configuration
	fltrDelay   int

	sTestsFltr = []func(t *testing.T){
		testV1FltrLoadConfig,
		testV1FltrInitDataDb,
		testV1FltrResetStorDb,
		testV1FltrStartEngine,
		testV1FltrRpcConn,
		testV1FltrLoadTarrifPlans,
		testV1FltrAddStats,
		testV1FltrPupulateThreshold,
		testV1FltrGetThresholdForEvent,
		testV1FltrGetThresholdForEvent2,
		testV1FltrPopulateResources,
		testV1FltrAccounts,
		testV1FltrStopEngine,
	}
)

// Test start here
func TestFltrIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		fltrConfDIR = "filters_internal"
	case utils.MetaMySQL:
		fltrConfDIR = "filters_mysql"
	case utils.MetaMongo:
		fltrConfDIR = "filters_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsFltr {
		t.Run(fltrConfDIR, stest)
	}
}

func testV1FltrLoadConfig(t *testing.T) {
	var err error
	fltrCfgPath = path.Join(*dataDir, "conf", "samples", fltrConfDIR)
	if *encoding == utils.MetaGOB {
		cdrsCfgPath = path.Join(*dataDir, "conf", "samples", fltrConfDIR+"_gob")
	}
	if fltrCfg, err = config.NewCGRConfigFromPath(fltrCfgPath); err != nil {
		t.Error(err)
	}
	fltrDelay = 1000
}

func testV1FltrInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrCfgPath, fltrDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrRpcConn(t *testing.T) {
	var err error
	fltrRpc, err = newRPCClient(fltrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FltrLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := fltrRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1FltrAddStats(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       10.0,
			},
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1001",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       10.5,
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(5 * time.Second),
			utils.COST:       12.5,
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(6 * time.Second),
			utils.COST:       17.5,
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_3"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.Account:    "1003",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       12.5,
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       12.5,
			utils.PDD:        time.Duration(12 * time.Second),
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(15 * time.Second),
			utils.COST:       15.5,
			utils.PDD:        time.Duration(15 * time.Second),
		},
	}
	if err := fltrRpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1FltrPupulateThreshold(t *testing.T) {
	//Add a filter of type *stats and check if acd metric is minim 10 ( greater than 10)
	//we expect that acd from Stat_1 to be 11 so the filter should pass (11 > 10)
	filter := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Stats1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*stats.Stat_1.*acd",
					Values:  []string{"10.0"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "LOG", Actions: []*utils.TPAction{
		{Identifier: utils.LOG},
	}}
	if err := fltrRpc.Call(utils.APIerSv2SetActions, attrsAA, &result); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", result)
	}
	time.Sleep(10 * time.Millisecond)

	//Add a threshold with filter from above and an inline filter for Account 1010
	tPrfl := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Stats1",
			FilterIDs: []string{"FLTR_TH_Stats1", "*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  time.Duration(1 * time.Millisecond),
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}
}

func testV1FltrGetThresholdForEvent(t *testing.T) {
	// check the event
	tEv := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1010"},
		},
	}
	var ids []string
	eIDs := []string{"TH_Stats1"}
	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
}

func testV1FltrGetThresholdForEvent2(t *testing.T) {
	//Add a filter of type *stats and check if acd metric is maximum 10 ( lower than 10)
	//we expect that acd from Stat_1 to be 11 so the filter should not pass (11 > 10)
	filter := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Stats1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*lt",
					Element: "~*stats.Stat_1.*acd",
					Values:  []string{"10.0"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//update the threshold with new filter
	tPrfl := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Stats1",
			FilterIDs: []string{"FLTR_TH_Stats1", "*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  time.Duration(1 * time.Millisecond),
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
		},
	}
	if err := fltrRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	tEv := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1010"},
		},
	}
	var ids []string
	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrPopulateResources(t *testing.T) {
	//create a resourceProfile
	rlsConfig := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResTest",
		UsageTTL:          time.Duration(1) * time.Minute,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}

	var result string
	if err := fltrRpc.Call(utils.APIerSv1SetResourceProfile, &v1.ResourceWithCache{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ResourceProfile
	if err := fltrRpc.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}

	// Allocate 3 units for resource ResTest
	argsRU := utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				"Account":     "3001",
				"Destination": "3002"},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Units:   3,
	}
	if err := fltrRpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	}

	//we allocate 3 units to resource and add a filter for Usages > 2
	//should match (3>2)
	filter := v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Resource",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*resources.ResTest.TotalUsage",
					Values:  []string{"2.0"},
				},
			},
		},
	}

	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	tPrfl := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_ResTest",
			FilterIDs: []string{"FLTR_TH_Resource", "*string:~*req.Account:2020"},
			MaxHits:   -1,
			MinSleep:  time.Duration(1 * time.Millisecond),
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}

	// check the event
	tEv := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "2020"},
		},
	}
	var ids []string
	eIDs := []string{"TH_ResTest"}
	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

	//change the filter
	//we allocate 3 units to resource and add a filter for Usages < 2
	//should fail (3<2)
	filter.Filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_Resource",
		Rules: []*engine.FilterRule{
			{
				Type:    "*lt",
				Element: "~*resources.ResTest.TotalUsage",
				Values:  []string{"2.0"},
			},
		},
	}

	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//Overwrite the threshold
	if err := fltrRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//expect NotFound error because filter doesn't match
	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrAccounts(t *testing.T) {
	var resp string
	if err := fltrRpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	// Add a filter with fieldName taken value from account 1001
	// and check if *monetary balance is minim 9 ( greater than 9)
	// we expect that the balance to be 10 so the filter should pass (10 > 9)
	filter := v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Accounts",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*accounts.1001.BalanceMap.*monetary[0].Value",
					Values:  []string{"9"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// Add a log action
	attrsAA := &utils.AttrSetActions{ActionsId: "LOG", Actions: []*utils.TPAction{
		{Identifier: utils.LOG},
	}}
	if err := fltrRpc.Call(utils.APIerSv2SetActions, attrsAA, &result); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", result)
	}
	time.Sleep(10 * time.Millisecond)
	//Add a threshold with filter from above and an inline filter for Account 1010
	tPrfl := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Account",
			FilterIDs: []string{"FLTR_TH_Accounts", "*string:~*req.Account:1001"},
			MaxHits:   -1,
			MinSleep:  time.Duration(1 * time.Millisecond),
			Weight:    90.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}

	tEv := &engine.ArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account: "1001"},
		},
	}
	var ids []string
	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"TH_Account"}) {
		t.Error("Unexpected reply returned", ids)
	}

	// update the filter
	// Add a filter with fieldName taken value from account 1001
	// and check if *monetary balance is is minim 11 ( greater than 11)
	// we expect that the balance to be 10 so the filter should not pass (10 > 11)
	filter.Filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_Accounts",
		Rules: []*engine.FilterRule{
			{
				Type:    "*gt",
				Element: "~*accounts.1001.BalanceMap.*monetary[0].Value",
				Values:  []string{"11"},
			},
		},
	}

	if err := fltrRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := fltrRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
