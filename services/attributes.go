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

// NewAttributeService returns the Attribute Service
func NewAttributeService(cfg *config.CGRConfig, dm *DataDBService,
	cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	server *utils.Server, internalChan chan rpcclient.ClientConnector) servmanager.Service {
	return &AttributeService{
		connChan:    internalChan,
		cfg:         cfg,
		dm:          dm,
		cacheS:      cacheS,
		filterSChan: filterSChan,
		server:      server,
	}
}

// AttributeService implements Service interface
type AttributeService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	cacheS      *engine.CacheS
	filterSChan chan *engine.FilterS
	server      *utils.Server

	attrS    *engine.AttributeService
	rpc      *v1.AttributeSv1
	connChan chan rpcclient.ClientConnector
}

// Start should handle the sercive start
func (attrS *AttributeService) Start() (err error) {
	if attrS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	<-attrS.cacheS.GetPrecacheChannel(utils.CacheAttributeProfiles)
	<-attrS.cacheS.GetPrecacheChannel(utils.CacheAttributeFilterIndexes)

	filterS := <-attrS.filterSChan
	attrS.filterSChan <- filterS
	dbchan := attrS.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	attrS.Lock()
	defer attrS.Unlock()
	attrS.attrS, err = engine.NewAttributeService(datadb, filterS, attrS.cfg)
	if err != nil {
		utils.Logger.Crit(
			fmt.Sprintf("<%s> Could not init, error: %s",
				utils.AttributeS, err.Error()))
		return
	}
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.AttributeS))
	attrS.rpc = v1.NewAttributeSv1(attrS.attrS)
	if !attrS.cfg.DispatcherSCfg().Enabled {
		attrS.server.RpcRegister(attrS.rpc)
	}
	attrS.connChan <- attrS.rpc
	return
}

// Reload handles the change of config
func (attrS *AttributeService) Reload() (err error) {
	return // for the momment nothing to reload
}

// Shutdown stops the service
func (attrS *AttributeService) Shutdown() (err error) {
	attrS.Lock()
	defer attrS.Unlock()
	if err = attrS.attrS.Shutdown(); err != nil {
		return
	}
	attrS.attrS = nil
	attrS.rpc = nil
	<-attrS.connChan
	return
}

// IsRunning returns if the service is running
func (attrS *AttributeService) IsRunning() bool {
	attrS.RLock()
	defer attrS.RUnlock()
	return attrS != nil && attrS.attrS != nil
}

// ServiceName returns the service name
func (attrS *AttributeService) ServiceName() string {
	return utils.AttributeS
}

// ShouldRun returns if the service should be running
func (attrS *AttributeService) ShouldRun() bool {
	return attrS.cfg.AttributeSCfg().Enabled
}
