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

package engine

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/guardian"
	"github.com/Omnitouch/cgrates/utils"
)

// NewStatService initializes a StatService
func NewStatService(dm *DataManager, cgrcfg *config.CGRConfig,
	filterS *FilterS, connMgr *ConnManager) (ss *StatService, err error) {

	return &StatService{
		dm:               dm,
		connMgr:          connMgr,
		filterS:          filterS,
		cgrcfg:           cgrcfg,
		storedStatQueues: make(utils.StringMap),
		loopStoped:       make(chan struct{}),
		stopBackup:       make(chan struct{})}, nil
}

// StatService builds stats for events
type StatService struct {
	dm               *DataManager
	connMgr          *ConnManager
	filterS          *FilterS
	cgrcfg           *config.CGRConfig
	loopStoped       chan struct{}
	stopBackup       chan struct{}
	storedStatQueues utils.StringMap // keep a record of stats which need saving, map[statsTenantID]bool
	ssqMux           sync.RWMutex    // protects storedStatQueues
}

// ListenAndServe loops keeps the service alive
func (sS *StatService) ListenAndServe(exitChan chan bool) error {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.StatS))
	go sS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (sS *StatService) Shutdown() error {
	utils.Logger.Info("<StatS> service shutdown initialized")
	close(sS.stopBackup)
	sS.storeStats()
	utils.Logger.Info("<StatS> service shutdown complete")
	return nil
}

// runBackup will regularly store resources changed to dataDB
func (sS *StatService) runBackup() {
	storeInterval := sS.cgrcfg.StatSCfg().StoreInterval
	if storeInterval <= 0 {
		sS.loopStoped <- struct{}{}
		return
	}
	for {
		sS.storeStats()
		select {
		case <-sS.stopBackup:
			sS.loopStoped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeResources represents one task of complete backup
func (sS *StatService) storeStats() {
	var failedSqIDs []string
	for { // don't stop untill we store all dirty statQueues
		sS.ssqMux.Lock()
		sID := sS.storedStatQueues.GetOne()
		if sID != "" {
			delete(sS.storedStatQueues, sID)
		}
		sS.ssqMux.Unlock()
		if sID == "" {
			break // no more keys, backup completed
		}
		lkID := utils.StatQueuePrefix + sID
		guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
			if sqIf, ok := Cache.Get(utils.CacheStatQueues, sID); !ok || sqIf == nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> failed retrieving from cache stat queue with ID: %s",
						utils.StatService, sID))
			} else if err := sS.StoreStatQueue(sqIf.(*StatQueue)); err != nil {
				failedSqIDs = append(failedSqIDs, sID) // record failure so we can schedule it for next backup
			}
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkID)
		// randomize the CPU load and give up thread control
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
	}
	if len(failedSqIDs) != 0 { // there were errors on save, schedule the keys for next backup
		sS.ssqMux.Lock()
		for _, sqID := range failedSqIDs {
			sS.storedStatQueues[sqID] = true
		}
		sS.ssqMux.Unlock()
	}
}

// StoreStatQueue stores the statQueue in DB and corrects dirty flag
func (sS *StatService) StoreStatQueue(sq *StatQueue) (err error) {
	if sq.dirty == nil || !*sq.dirty {
		return
	}
	if err = sS.dm.SetStatQueue(sq); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed saving StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if err = sS.dm.CacheDataFromDB(utils.StatQueuePrefix, []string{sq.TenantID()}, true); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed caching StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	*sq.dirty = false
	return
}

// matchingStatQueuesForEvent returns ordered list of matching resources which are active by the time of the call
func (sS *StatService) matchingStatQueuesForEvent(args *StatsArgsProcessEvent) (sqs StatQueues, err error) {
	matchingSQs := make(map[string]*StatQueue)
	var sqIDs []string
	if len(args.StatIDs) != 0 {
		sqIDs = args.StatIDs
	} else {
		mapIDs, err := MatchingItemIDsForEvent(args.Event,
			sS.cgrcfg.StatSCfg().StringIndexedFields,
			sS.cgrcfg.StatSCfg().PrefixIndexedFields,
			sS.dm, utils.CacheStatFilterIndexes, args.Tenant,
			sS.cgrcfg.StatSCfg().IndexedSelects,
			sS.cgrcfg.StatSCfg().NestedFields,
		)
		if err != nil {
			return nil, err
		}
		sqIDs = mapIDs.Slice()
	}
	evNm := utils.MapStorage{utils.MetaReq: args.Event}
	for _, sqID := range sqIDs {
		sqPrfl, err := sS.dm.GetStatQueueProfile(args.Tenant, sqID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if sqPrfl.ActivationInterval != nil && args.Time != nil &&
			!sqPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			continue
		}
		if pass, err := sS.filterS.Pass(args.Tenant, sqPrfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		var sq *StatQueue
		lkID := utils.StatQueuePrefix + utils.ConcatenatedKey(sqPrfl.Tenant, sqPrfl.ID)
		guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
			sq, err = sS.dm.GetStatQueue(sqPrfl.Tenant, sqPrfl.ID, true, true, "")
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkID)
		if err != nil {
			return nil, err
		}
		if sqPrfl.Stored && sq.dirty == nil {
			sq.dirty = utils.BoolPointer(false)
		}
		if sqPrfl.TTL >= 0 {
			sq.ttl = utils.DurationPointer(sqPrfl.TTL)
		}
		sq.sqPrfl = sqPrfl
		matchingSQs[sqPrfl.ID] = sq
	}
	// All good, convert from Map to Slice so we can sort
	sqs = make(StatQueues, len(matchingSQs))
	i := 0
	for _, s := range matchingSQs {
		sqs[i] = s
		i++
	}
	sqs.Sort()
	for i, s := range sqs {
		if s.sqPrfl.Blocker { // blocker will stop processing
			sqs = sqs[:i+1]
			break
		}
	}
	return
}

// Call implements rpcclient.ClientConnector interface for internal RPC
// here for cases when passing StatsService as rpccclient.RpcClientConnection
func (ss *StatService) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(ss, serviceMethod, args, reply)
}

type StatsArgsProcessEvent struct {
	StatIDs []string
	*utils.CGREvent
	*utils.ArgDispatcher
}

// processEvent processes a new event, dispatching to matching queues
// queues matching are also cached to speed up
func (sS *StatService) processEvent(args *StatsArgsProcessEvent) (statQueueIDs []string, err error) {
	matchSQs, err := sS.matchingStatQueuesForEvent(args)
	if err != nil {
		return nil, err
	}
	if len(matchSQs) == 0 {
		return nil, utils.ErrNotFound
	}
	var stsIDs []string
	var withErrors bool
	for _, sq := range matchSQs {
		stsIDs = append(stsIDs, sq.ID)
		lkID := utils.StatQueuePrefix + sq.TenantID()
		guardian.Guardian.Guard(func() (gRes interface{}, gErr error) {
			err = sq.ProcessEvent(args.CGREvent, sS.filterS)
			return
		}, config.CgrConfig().GeneralCfg().LockingTimeout, lkID)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Queue: %s, ignoring event: %s, error: %s",
					sq.TenantID(), args.TenantID(), err.Error()))
			withErrors = true
		}
		if sS.cgrcfg.StatSCfg().StoreInterval != 0 && sq.dirty != nil { // don't save
			if sS.cgrcfg.StatSCfg().StoreInterval == -1 {
				*sq.dirty = true
				sS.StoreStatQueue(sq)
			} else {
				*sq.dirty = true // mark it to be saved
				sS.ssqMux.Lock()
				sS.storedStatQueues[sq.TenantID()] = true
				sS.ssqMux.Unlock()
			}
		}
		if len(sS.cgrcfg.StatSCfg().ThresholdSConns) != 0 {
			var thIDs []string
			if len(sq.sqPrfl.ThresholdIDs) != 0 {
				if len(sq.sqPrfl.ThresholdIDs) == 1 && sq.sqPrfl.ThresholdIDs[0] == utils.META_NONE {
					continue
				}
				thIDs = sq.sqPrfl.ThresholdIDs
			}
			thEv := &ArgsProcessEvent{
				ThresholdIDs: thIDs,
				CGREvent: &utils.CGREvent{
					Tenant: sq.Tenant,
					ID:     utils.GenUUID(),
					Event: map[string]interface{}{
						utils.EventType: utils.StatUpdate,
						utils.StatID:    sq.ID,
					},
				},
				ArgDispatcher: args.ArgDispatcher,
			}
			for metricID, metric := range sq.SQMetrics {
				thEv.Event[metricID] = metric.GetValue()
			}
			var tIDs []string
			if err := sS.connMgr.Call(sS.cgrcfg.StatSCfg().ThresholdSConns, nil,
				utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<StatS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
				withErrors = true
			}
		}
	}
	if len(stsIDs) != 0 {
		statQueueIDs = append(statQueueIDs, stsIDs...)
	} else {
		statQueueIDs = []string{}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatService) V1ProcessEvent(args *StatsArgsProcessEvent, reply *[]string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.Tenant, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if ids, err := sS.processEvent(args); err != nil {
		return err
	} else {
		*reply = ids
	}
	return
}

// V1StatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatService) V1GetStatQueuesForEvent(args *StatsArgsProcessEvent, reply *[]string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.Tenant, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	var sQs StatQueues
	if sQs, err = sS.matchingStatQueuesForEvent(args); err != nil {
		return
	} else {
		ids := make([]string, len(sQs))
		for i, sq := range sQs {
			ids[i] = sq.ID
		}
		*reply = ids
	}
	return
}

// V1GetStatQueue returns a StatQueue object
func (sS *StatService) V1GetStatQueue(args *utils.TenantIDWithArgDispatcher, reply *StatQueue) (err error) {
	if sq, err := sS.dm.GetStatQueue(args.Tenant, args.ID, true, true, ""); err != nil {
		return err
	} else {
		*reply = *sq
	}
	return
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (sS *StatService) V1GetQueueStringMetrics(args *utils.TenantID, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.Tenant, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sq, err := sS.dm.GetStatQueue(args.Tenant, args.ID, true, true, "")
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	sq.RLock()
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue("")
	}
	sq.RUnlock()
	*reply = metrics
	return
}

// V1GetFloatMetrics returns the metrics as float64 values
func (sS *StatService) V1GetQueueFloatMetrics(args *utils.TenantID, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.Tenant, utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sq, err := sS.dm.GetStatQueue(args.Tenant, args.ID, true, true, "")
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	sq.RLock()
	metrics := make(map[string]float64, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetFloat64Value()
	}
	sq.RUnlock()
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatService) V1GetQueueIDs(tenant string, qIDs *[]string) (err error) {
	prfx := utils.StatQueuePrefix + tenant + ":"
	keys, err := sS.dm.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*qIDs = retIDs
	return
}

// Reload stops the backupLoop and restarts it
func (sS *StatService) Reload() {
	close(sS.stopBackup)
	<-sS.loopStoped // wait until the loop is done
	sS.stopBackup = make(chan struct{})
	go sS.runBackup()
}

// StartLoop starsS the gorutine with the backup loop
func (sS *StatService) StartLoop() {
	go sS.runBackup()
}
