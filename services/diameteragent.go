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

	"github.com/Omnitouch/cgrates/agents"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
)

// NewDiameterAgent returns the Diameter Agent
func NewDiameterAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &DiameterAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// DiameterAgent implements Agent interface
type DiameterAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	exitChan    chan bool

	da      *agents.DiameterAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (da *DiameterAgent) Start() (err error) {
	if da.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-da.filterSChan
	da.filterSChan <- filterS

	da.Lock()
	defer da.Unlock()

	da.da, err = agents.NewDiameterAgent(da.cfg, filterS, da.connMgr)
	if err != nil {
		utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
			utils.DiameterAgent, err))
		return
	}

	go func() {
		if err = da.da.ListenAndServe(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!",
				utils.DiameterAgent, err))
		}
		da.exitChan <- true
	}()
	return
}

// Reload handles the change of config
func (da *DiameterAgent) Reload() (err error) {
	return
}

// Shutdown stops the service
func (da *DiameterAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (da *DiameterAgent) IsRunning() bool {
	da.RLock()
	defer da.RUnlock()
	return da != nil && da.da != nil
}

// ServiceName returns the service name
func (da *DiameterAgent) ServiceName() string {
	return utils.DiameterAgent
}

// ShouldRun returns if the service should be running
func (da *DiameterAgent) ShouldRun() bool {
	return da.cfg.DiameterAgentCfg().Enabled
}
