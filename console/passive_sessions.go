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
	"github.com/Omnitouch/cgrates/sessions"
	"github.com/Omnitouch/cgrates/utils"
)

func init() {
	c := &CmdPassiveSessions{
		name:      "passive_sessions",
		rpcMethod: utils.SessionSv1GetPassiveSessions,
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdPassiveSessions struct {
	name      string
	rpcMethod string
	rpcParams interface{}
	*CommandExecuter
}

func (self *CmdPassiveSessions) Name() string {
	return self.name
}

func (self *CmdPassiveSessions) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdPassiveSessions) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &utils.SessionFilter{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdPassiveSessions) PostprocessRpcParams() error {
	param := self.rpcParams.(*utils.SessionFilter)
	self.rpcParams = param
	return nil
}

func (self *CmdPassiveSessions) RpcResult() interface{} {
	var sessions *[]*sessions.ExternalSession
	return &sessions
}

func (self *CmdPassiveSessions) GetFormatedResult(result interface{}) string {
	return GetFormatedSliceResult(result, map[string]struct{}{
		"Usage":         struct{}{},
		"DurationIndex": struct{}{},
		"MaxRateUnit":   struct{}{},
		"DebitInterval": struct{}{},
	})
}
