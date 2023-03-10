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

package console

import (
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func init() {
	c := &CmdGetThresholdProfile{
		name:      "threshold_profile",
		rpcMethod: utils.APIerSv1GetThresholdProfile,
		rpcParams: &utils.TenantIDWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

type CmdGetThresholdProfile struct {
	name      string
	rpcMethod string
	rpcParams *utils.TenantIDWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetThresholdProfile) Name() string {
	return self.name
}

func (self *CmdGetThresholdProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetThresholdProfile) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.TenantIDWithArgDispatcher{
			TenantID:      new(utils.TenantID),
			ArgDispatcher: new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdGetThresholdProfile) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetThresholdProfile) RpcResult() interface{} {
	var atr engine.ThresholdProfile
	return &atr
}

func (self *CmdGetThresholdProfile) GetFormatedResult(result interface{}) string {
	return GetFormatedResult(result, map[string]struct{}{
		"MinSleep": struct{}{},
	})
}
