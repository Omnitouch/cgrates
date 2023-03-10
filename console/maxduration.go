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
	"encoding/json"
	"fmt"
	"time"

	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func init() {
	c := &CmdGetMaxDuration{
		name:       "maxduration",
		rpcMethod:  utils.ResponderGetMaxSessionTime,
		clientArgs: []string{"Category", "ToR", "Tenant", "Subject", "Account", "Destination", "TimeStart", "TimeEnd", "CallDuration", "FallbackSubject"},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetMaxDuration struct {
	name       string
	rpcMethod  string
	rpcParams  *engine.CallDescriptorWithArgDispatcher
	clientArgs []string
	*CommandExecuter
}

func (self *CmdGetMaxDuration) Name() string {
	return self.name
}

func (self *CmdGetMaxDuration) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetMaxDuration) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &engine.CallDescriptorWithArgDispatcher{
			CallDescriptor: new(engine.CallDescriptor),
			ArgDispatcher:  new(utils.ArgDispatcher),
		}
	}
	return self.rpcParams
}

func (self *CmdGetMaxDuration) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetMaxDuration) RpcResult() interface{} {
	var d time.Duration
	return &d
}

func (self *CmdGetMaxDuration) ClientArgs() []string {
	return self.clientArgs
}

func (self *CmdGetMaxDuration) GetFormatedResult(result interface{}) string {
	if tv, canCast := result.(*time.Duration); canCast {
		return fmt.Sprintf(`"%s"`, tv.String())
	}
	out, _ := json.MarshalIndent(result, "", " ")
	return string(out)
}
