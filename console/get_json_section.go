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
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/utils"
)

func init() {
	c := &CmdGetJSONConfig{
		name:      "get_json_section",
		rpcMethod: utils.ConfigSv1GetJSONSection,
		rpcParams: &config.StringWithArgDispatcher{},
	}
	commands[c.Name()] = c
	c.CommandExecuter = &CommandExecuter{c}
}

// Commander implementation
type CmdGetJSONConfig struct {
	name      string
	rpcMethod string
	rpcParams *config.StringWithArgDispatcher
	*CommandExecuter
}

func (self *CmdGetJSONConfig) Name() string {
	return self.name
}

func (self *CmdGetJSONConfig) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdGetJSONConfig) RpcParams(reset bool) interface{} {
	if reset || self.rpcParams == nil {
		self.rpcParams = &config.StringWithArgDispatcher{ArgDispatcher: new(utils.ArgDispatcher)}
	}
	return self.rpcParams
}

func (self *CmdGetJSONConfig) PostprocessRpcParams() error {
	return nil
}

func (self *CmdGetJSONConfig) RpcResult() interface{} {
	var s map[string]interface{}
	return &s
}
