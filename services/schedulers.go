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

	v1 "github.com/Omnitouch/cgrates/apier/v1"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/scheduler"
	"github.com/Omnitouch/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewSchedulerService returns the Scheduler Service
func NewSchedulerService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, fltrSChan chan *engine.FilterS,
	server *utils.Server, internalSchedulerrSChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager) *SchedulerService {
	return &SchedulerService{
		connChan:  internalSchedulerrSChan,
		cfg:       cfg,
		dm:        dm,
		cacheS:    cacheS,
		fltrSChan: fltrSChan,
		server:    server,
		connMgr:   connMgr,
	}
}

// SchedulerService implements Service interface
type SchedulerService struct {
	sync.RWMutex
	cfg       *config.CGRConfig
	dm        *DataDBService
	cacheS    *engine.CacheS
	fltrSChan chan *engine.FilterS
	server    *utils.Server

	schS     *scheduler.Scheduler
	rpc      *v1.SchedulerSv1
	connChan chan rpcclient.ClientConnector
	connMgr  *engine.ConnManager
}

// Start should handle the sercive start
func (schS *SchedulerService) Start() (err error) {
	if schS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-schS.cacheS.GetPrecacheChannel(utils.CacheActionPlans) // wait for ActionPlans to be cached

	fltrS := <-schS.fltrSChan
	schS.fltrSChan <- fltrS
	dbchan := schS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	schS.Lock()
	defer schS.Unlock()
	utils.Logger.Info("<ServiceManager> Starting CGRateS Scheduler.")
	schS.schS = scheduler.NewScheduler(datadb, schS.cfg, fltrS)
	go schS.schS.Loop()

	schS.rpc = v1.NewSchedulerSv1(schS.cfg)
	if !schS.cfg.DispatcherSCfg().Enabled {
		schS.server.RpcRegister(schS.rpc)
	}
	schS.connChan <- schS.rpc

	return
}

// Reload handles the change of config
func (schS *SchedulerService) Reload() (err error) {
	schS.Lock()
	schS.schS.Reload()
	schS.Unlock()
	return
}

// Shutdown stops the service
func (schS *SchedulerService) Shutdown() (err error) {
	schS.Lock()
	schS.schS.Shutdown()
	schS.schS = nil
	schS.rpc = nil
	<-schS.connChan
	schS.Unlock()
	return
}

// IsRunning returns if the service is running
func (schS *SchedulerService) IsRunning() bool {
	schS.RLock()
	defer schS.RUnlock()
	return schS != nil && schS.schS != nil
}

// ServiceName returns the service name
func (schS *SchedulerService) ServiceName() string {
	return utils.SchedulerS
}

// GetScheduler returns the Scheduler
func (schS *SchedulerService) GetScheduler() *scheduler.Scheduler {
	schS.RLock()
	defer schS.RUnlock()
	return schS.schS
}

// ShouldRun returns if the service should be running
func (schS *SchedulerService) ShouldRun() bool {
	return schS.cfg.SchedulerCfg().Enabled
}
