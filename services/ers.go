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

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/ers"
	"github.com/Omnitouch/cgrates/servmanager"
	"github.com/Omnitouch/cgrates/utils"
)

// NewEventReaderService returns the EventReader Service
func NewEventReaderService(cfg *config.CGRConfig, filterSChan chan *engine.FilterS,
	exitChan chan bool, connMgr *engine.ConnManager) servmanager.Service {
	return &EventReaderService{
		rldChan:     make(chan struct{}, 1),
		cfg:         cfg,
		filterSChan: filterSChan,
		exitChan:    exitChan,
		connMgr:     connMgr,
	}
}

// EventReaderService implements Service interface
type EventReaderService struct {
	sync.RWMutex
	cfg         *config.CGRConfig
	filterSChan chan *engine.FilterS
	exitChan    chan bool

	ers      *ers.ERService
	rldChan  chan struct{}
	stopChan chan struct{}
	connMgr  *engine.ConnManager
}

// Start should handle the sercive start
func (erS *EventReaderService) Start() (err error) {
	if erS.IsRunning() {
		return utils.ErrServiceAlreadyRunning
	}

	erS.Lock()
	defer erS.Unlock()

	filterS := <-erS.filterSChan
	erS.filterSChan <- filterS

	// remake tht stop chan
	erS.stopChan = make(chan struct{}, 1)

	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ERs))

	// build the service
	erS.ers = ers.NewERService(erS.cfg, filterS, erS.stopChan, erS.connMgr)
	go func(ers *ers.ERService, rldChan chan struct{}) {
		if err := ers.ListenAndServe(rldChan); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: <%s>", utils.ERs, err.Error()))
			erS.exitChan <- true
		}
	}(erS.ers, erS.rldChan)
	return
}

// Reload handles the change of config
func (erS *EventReaderService) Reload() (err error) {
	erS.RLock()
	erS.rldChan <- struct{}{}
	erS.RUnlock()
	return
}

// Shutdown stops the service
func (erS *EventReaderService) Shutdown() (err error) {
	erS.Lock()
	close(erS.stopChan)
	erS.ers = nil
	erS.Unlock()
	return
}

// IsRunning returns if the service is running
func (erS *EventReaderService) IsRunning() bool {
	erS.RLock()
	defer erS.RUnlock()
	return erS != nil && erS.ers != nil
}

// ServiceName returns the service name
func (erS *EventReaderService) ServiceName() string {
	return utils.ERs
}

// ShouldRun returns if the service should be running
func (erS *EventReaderService) ShouldRun() bool {
	return erS.cfg.ERsCfg().Enabled
}
