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

	v1 "github.com/Omnitouch/cgrates/apier/v1"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewThresholdService returns the Threshold Service
func NewThresholdService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalThresholdSChan chan rpcclient.ClientConnector) servmanager.Service {
	return &ThresholdService{
		connChan:    internalThresholdSChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
	}
}

// ThresholdService implements Service interface
type ThresholdService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *utils.Server

	thrs     *engine.ThresholdService
	rpc      *v1.ThresholdSv1
	connChan chan rpcclient.ClientConnector
}

// Start should handle the sercive start
func (thrs *ThresholdService) Start() (err error) {
	if thrs.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholdProfiles)
	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholds)
	<-thrs.cacheS.GetPrecacheChannel(utils.CacheThresholdFilterIndexes)

	filterS := <-thrs.filterSChan
	thrs.filterSChan <- filterS
	dbchan := thrs.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	thrs.Lock()
	defer thrs.Unlock()
	thrs.thrs, err = engine.NewThresholdService(datadb, thrs.cfg, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.ThresholdS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ThresholdS))
	thrs.thrs.StartLoop()
	thrs.rpc = v1.NewThresholdSv1(thrs.thrs)
	if !thrs.cfg.DispatcherSCfg().Enabled {
		thrs.server.RpcRegister(thrs.rpc)
	}
	thrs.connChan <- thrs.rpc
	return
}

// Reload handles the change of config
func (thrs *ThresholdService) Reload() (err error) {
	thrs.Lock()
	thrs.thrs.Reload()
	thrs.Unlock()
	return
}

// Shutdown stops the service
func (thrs *ThresholdService) Shutdown() (err error) {
	thrs.Lock()
	defer thrs.Unlock()
	if err = thrs.thrs.Shutdown(); err != nil {
		return
	}
	thrs.thrs = nil
	thrs.rpc = nil
	<-thrs.connChan
	return
}

// IsRunning returns if the service is running
func (thrs *ThresholdService) IsRunning() bool {
	thrs.RLock()
	defer thrs.RUnlock()
	return thrs != nil && thrs.thrs != nil
}

// ServiceName returns the service name
func (thrs *ThresholdService) ServiceName() string {
	return utils.ThresholdS
}

// ShouldRun returns if the service should be running
func (thrs *ThresholdService) ShouldRun() bool {
	return thrs.cfg.ThresholdSCfg().Enabled
}
