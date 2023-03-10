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

package v2

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	v1 "github.com/Omnitouch/cgrates/apier/v1"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

var (
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *rpc.Client
	alsPrfDataDir   = "/usr/share/cgrates"
	alsPrf          *engine.AttributeProfile
	alsPrfConfigDIR string //run tests for specific configuration

	sTestsAlsPrf = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSSetAlsPrf,
		testAttributeSUpdateAlsPrf,
		testAttributeSKillEngine,
	}
)

//Test start here
func TestAttributeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		alsPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	alsPrfCfgPath = path.Join(alsPrfDataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromPath(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	alsPrfCfg.DataFolderPath = alsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(alsPrfCfg)
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = newRPCClient(alsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSSetAlsPrf(t *testing.T) {
	extAlsPrf := &AttributeWithCache{
		ExternalAttributeProfile: &engine.ExternalAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: "1001",
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv2SetAttributeProfile, extAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ExternalAttribute"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testAttributeSUpdateAlsPrf(t *testing.T) {
	extAlsPrf := &AttributeWithCache{
		ExternalAttributeProfile: &engine.ExternalAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: "1001",
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Value: "~*req.Account",
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv2SetAttributeProfile, extAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ExternalAttribute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Value: config.NewRSRParsersMustCompile("~*req.Account", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	sort.Strings(alsPrf.AttributeProfile.Contexts)
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ExternalAttribute"}}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.Contexts)
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
}

func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
