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

package services

import (
	"fmt"
	"sync"

	"github.com/Omnitouch/cgrates/engine"

	"github.com/Omnitouch/cgrates/agents"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
)

// NewAsteriskAgent returns the Asterisk Agent
func NewAsteriskAgent(cfg *config.CGRConfig,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &AsteriskAgent{
		cfg:      cfg,
		exitChan: exitChan,
		connMgr:  connMgr,
	}
}

// AsteriskAgent implements Agent interface
type AsteriskAgent struct {
	sync.RWMutex
	cfg      *config.CGRConfig
	exitChan chan bool

	smas    []*agents.AsteriskAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (ast *AsteriskAgent) Start() (err error) {
	if ast.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	ast.Lock()
	defer ast.Unlock()

	listenAndServe := func(sma *agents.AsteriskAgent, exitChan chan bool) {
		if err = sma.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> runtime error: %s!", utils.AsteriskAgent, err))
		}
		exitChan <- true
	}
	ast.smas = make([]*agents.AsteriskAgent, len(ast.cfg.AsteriskAgentCfg().AsteriskConns))
	for connIdx := range ast.cfg.AsteriskAgentCfg().AsteriskConns { // Instantiate connections towards asterisk servers
		if ast.smas[connIdx], err = agents.NewAsteriskAgent(ast.cfg, connIdx, ast.connMgr); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.AsteriskAgent, err))
			return
		}
		go listenAndServe(ast.smas[connIdx], ast.exitChan)
	}
	return
}

// Reload handles the change of config
func (ast *AsteriskAgent) Reload() (err error) {
	return
}

// Shutdown stops the service
func (ast *AsteriskAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ast *AsteriskAgent) IsRunning() bool {
	ast.RLock()
	defer ast.RUnlock()
	return ast != nil && ast.smas != nil
}

// ServiceName returns the service name
func (ast *AsteriskAgent) ServiceName() string {
	return utils.AsteriskAgent
}

// ShouldRun returns if the service should be running
func (ast *AsteriskAgent) ShouldRun() bool {
	return ast.cfg.AsteriskAgentCfg().Enabled
}
