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
	"runtime"
	"sync"

	v1 "github.com/Omnitouch/cgrates/apier/v1"
	v2 "github.com/Omnitouch/cgrates/apier/v2"
	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewCDRServer returns the CDR Server
func NewCDRServer(cfg *config.CGRConfig, dm *DataDBService,
	storDB *StorDBService, filterSChan chan *engine.FilterS,
	server *utils.Server, internalCDRServerChan chan rpcclient.ClientConnector,
	connMgr *engine.ConnManager) servmanager.Service {
	return &CDRServer{
		connChan:    internalCDRServerChan,
		cfg:         cfg,
		dm:          dm,
		storDB:      storDB,
		filterSChan: filterSChan,
		server:      server,
		connMgr:     connMgr,
	}
}

// CDRServer implements Service interface
type CDRServer struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	dm          *DataDBService
	storDB      *StorDBService
	filterSChan chan *engine.FilterS
	server      *utils.Server

	cdrS     *engine.CDRServer
	rpcv1    *v1.CDRsV1
	rpcv2    *v2.CDRsV2
	connChan chan rpcclient.ClientConnector
	connMgr  *engine.ConnManager

	syncStop chan struct{}
	// storDBChan chan engine.StorDB
}

// Start should handle the sercive start
func (cdrService *CDRServer) Start() (err error) {
	if cdrService.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.CDRs))

	filterS := <-cdrService.filterSChan
	cdrService.filterSChan <- filterS
	dbchan := cdrService.dm.GetDMChan()
	datadb := <-dbchan
	dbchan <- datadb

	cdrService.Lock()
	defer cdrService.Unlock()

	storDBChan := make(chan engine.StorDB, 1)
	cdrService.syncStop = make(chan struct{})
	cdrService.storDB.RegisterSyncChan(storDBChan)

	cdrService.cdrS = engine.NewCDRServer(cdrService.cfg, storDBChan, datadb, filterS, cdrService.connMgr)
	go func(cdrS *engine.CDRServer, stopChan chan struct{}) {
		if err := cdrS.ListenAndServe(stopChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.CDRServer, err.Error()))
			// erS.exitChan <- true
		}
	}(cdrService.cdrS, cdrService.syncStop)
	runtime.Gosched()
	utils.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrService.cdrS.RegisterHandlersToServer(cdrService.server)
	utils.Logger.Info("Registering CDRS RPC service.")
	cdrService.rpcv1 = v1.NewCDRsV1(cdrService.cdrS)
	cdrService.rpcv2 = &v2.CDRsV2{CDRsV1: *cdrService.rpcv1}
	if !cdrService.cfg.DispatcherSCfg().Enabled {
		cdrService.server.RpcRegister(cdrService.rpcv1)
		cdrService.server.RpcRegister(cdrService.rpcv2)
		// Make the cdr server available for internal communication
		cdrService.server.RpcRegister(cdrService.cdrS) // register CdrServer for internal usage (TODO: refactor this)
	}
	cdrService.connChan <- cdrService.cdrS // Signal that cdrS is operational
	return
}

// Reload handles the change of config
func (cdrService *CDRServer) Reload() (err error) {
	return
}

// Shutdown stops the service
func (cdrService *CDRServer) Shutdown() (err error) {
	cdrService.Lock()
	close(cdrService.syncStop)
	cdrService.cdrS = nil
	cdrService.rpcv1 = nil
	cdrService.rpcv2 = nil
	<-cdrService.connChan
	cdrService.Unlock()
	return
}

// IsRunning returns if the service is running
func (cdrService *CDRServer) IsRunning() bool {
	cdrService.RLock()
	defer cdrService.RUnlock()
	return cdrService != nil && cdrService.cdrS != nil
}

// ServiceName returns the service name
func (cdrService *CDRServer) ServiceName() string {
	return utils.CDRServer
}

// ShouldRun returns if the service should be running
func (cdrService *CDRServer) ShouldRun() bool {
	return cdrService.cfg.CdrsCfg().Enabled
}
