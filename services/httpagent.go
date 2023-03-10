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
	"sync"

	"github.com/Omnitouch/cgrates/agents"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
)

// NewHTTPAgent returns the HTTP Agent
func NewHTTPAgent(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	server *utils.Server, connMgr *engine.ConnManager) servmanager.Service {
	return &HTTPAgent{
		cfg:         cfg,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
	}
}

// HTTPAgent implements Agent interface
type HTTPAgent struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	server      *utils.Server

	ha      *agents.HTTPAgent
	connMgr *engine.ConnManager
}

// Start should handle the sercive start
func (ha *HTTPAgent) Start() (err error) {
	if ha.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	filterS := <-ha.filterSChan
	ha.filterSChan <- filterS

	ha.Lock()
	defer ha.Unlock()
	utils.Logger.Info("Starting HTTP agent")
	for _, agntCfg := range ha.cfg.HttpAgentCfg() {
		ha.server.RegisterHttpHandler(agntCfg.Url,
			agents.NewHTTPAgent(ha.connMgr, agntCfg.SessionSConns, filterS,
				ha.cfg.GeneralCfg().DefaultTenant, agntCfg.RequestPayload,
				agntCfg.ReplyPayload, agntCfg.RequestProcessors))
	}
	return
}

// Reload handles the change of config
func (ha *HTTPAgent) Reload() (err error) {
	return // no reload
}

// Shutdown stops the service
func (ha *HTTPAgent) Shutdown() (err error) {
	return // no shutdown for the momment
}

// IsRunning returns if the service is running
func (ha *HTTPAgent) IsRunning() bool {
	ha.RLock()
	defer ha.RUnlock()
	return ha != nil && ha.ha != nil
}

// ServiceName returns the service name
func (ha *HTTPAgent) ServiceName() string {
	return utils.HTTPAgent
}

// ShouldRun returns if the service should be running
func (ha *HTTPAgent) ShouldRun() bool {
	return len(ha.cfg.HttpAgentCfg()) != 0
}
