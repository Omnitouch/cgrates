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

package agents

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/sessions"
	"github.com/Omnitouch/cgrates/utils"
	"github.com/cgrates/fsock"
)

func NewFSsessions(fsAgentConfig *config.FsAgentCfg,
	timezone string, connMgr *engine.ConnManager) (fsa *FSsessions) {
	return &FSsessions{
		cfg:         fsAgentConfig,
		conns:       make([]*fsock.FSock, len(fsAgentConfig.EventSocketConns)),
		senderPools: make([]*fsock.FSockPool, len(fsAgentConfig.EventSocketConns)),
		timezone:    timezone,
		connMgr:     connMgr,
	}
}

// The freeswitch session manager type holding a buffer for the network connection
// and the active sessions
type FSsessions struct {
	cfg         *config.FsAgentCfg
	conns       []*fsock.FSock     // Keep the list here for connection management purposes
	senderPools []*fsock.FSockPool // Keep sender pools here
	timezone    string
	connMgr     *engine.ConnManager
}

func (sm *FSsessions) createHandlers() map[string][]func(string, int) {
	ca := func(body string, connIdx int) {
		sm.onChannelAnswer(
			NewFSEvent(body), connIdx)
	}
	ch := func(body string, connIdx int) {
		sm.onChannelHangupComplete(
			NewFSEvent(body), connIdx)
	}
	handlers := map[string][]func(string, int){
		"CHANNEL_ANSWER":          {ca},
		"CHANNEL_HANGUP_COMPLETE": {ch},
	}
	if sm.cfg.SubscribePark {
		cp := func(body string, connIdx int) {
			sm.onChannelPark(
				NewFSEvent(body), connIdx)
		}
		handlers["CHANNEL_PARK"] = []func(string, int){cp}
	}
	return handlers
}

// Sets the call timeout valid of starting of the call
func (sm *FSsessions) setMaxCallDuration(uuid string, connIdx int,
	maxDur time.Duration, destNr string) error {
	if len(sm.cfg.EmptyBalanceContext) != 0 {
		_, err := sm.conns[connIdx].SendApiCmd(
			fmt.Sprintf("uuid_setvar %s execute_on_answer sched_transfer +%d %s XML %s\n\n",
				uuid, int(maxDur.Seconds()), destNr, sm.cfg.EmptyBalanceContext))
		if err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
			return err
		}
		return nil
	}
	if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
		if _, err := sm.conns[connIdx].SendApiCmd(
			fmt.Sprintf("sched_broadcast +%d %s playback!manager_request::%s aleg\n\n",
				int(maxDur.Seconds()), uuid, sm.cfg.EmptyBalanceAnnFile)); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
			return err
		}
		return nil
	}
	_, err := sm.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s execute_on_answer sched_hangup +%d alloted_timeout\n\n",
			uuid, int(maxDur.Seconds())))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send sched_hangup command to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return err
	}
	return nil
}

// Sends the transfer command to unpark the call to freeswitch
func (sm *FSsessions) unparkCall(uuid string, connIdx int, call_dest_nb, notify string) (err error) {
	_, err = sm.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api notification to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return
	}
	if _, err = sm.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_transfer %s %s\n\n", uuid, call_dest_nb)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send unpark api call to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
	}
	return
}

func (sm *FSsessions) onChannelPark(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.META_NONE { // Not for us
		return
	}
	if connIdx >= len(sm.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(sm.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], sm.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	authArgs := fsev.V1AuthorizeArgs()
	authArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID
	var authReply sessions.V1AuthorizeReply
	if err := sm.connMgr.Call(sm.cfg.SessionSConns, sm, utils.SessionSv1AuthorizeEvent, authArgs, &authReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not authorize event %s, error: %s",
				utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		sm.unparkCall(fsev.GetUUID(), connIdx,
			fsev.GetCallDestNr(utils.MetaDefault), err.Error())
		return
	}
	if authReply.Attributes != nil {
		for _, fldName := range authReply.Attributes.AlteredFields {
			fldName = strings.TrimPrefix(fldName, utils.MetaReq+utils.NestingSep)
			if _, has := authReply.Attributes.CGREvent.Event[fldName]; !has {
				continue //maybe removed
			}
			if _, err := sm.conns[connIdx].SendApiCmd(
				fmt.Sprintf("uuid_setvar %s %s %s\n\n", fsev.GetUUID(), fldName,
					authReply.Attributes.CGREvent.Event[fldName])); err != nil {
				utils.Logger.Info(
					fmt.Sprintf("<%s> error %s setting channel variabile: %s",
						utils.FreeSWITCHAgent, err.Error(), fldName))
				sm.unparkCall(fsev.GetUUID(), connIdx,
					fsev.GetCallDestNr(utils.MetaDefault), err.Error())
				return
			}
		}
	}
	if authArgs.GetMaxUsage {
		if authReply.MaxUsage == 0 {
			sm.unparkCall(fsev.GetUUID(), connIdx,
				fsev.GetCallDestNr(utils.MetaDefault), utils.ErrInsufficientCredit.Error())
			return
		}
		sm.setMaxCallDuration(fsev.GetUUID(), connIdx,
			authReply.MaxUsage, fsev.GetCallDestNr(utils.MetaDefault))
	}
	if authReply.ResourceAllocation != nil {
		if _, err := sm.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), CGRResourceAllocation, *authReply.ResourceAllocation)); err != nil {
			utils.Logger.Info(
				fmt.Sprintf("<%s> error %s setting channel variabile: %s",
					utils.FreeSWITCHAgent, err.Error(), CGRResourceAllocation))
			sm.unparkCall(fsev.GetUUID(), connIdx,
				fsev.GetCallDestNr(utils.MetaDefault), err.Error())
			return
		}
	}
	if authReply.Suppliers != nil {
		fsArray := SliceAsFsArray(authReply.Suppliers.SuppliersWithParams())
		if _, err := sm.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_setvar %s %s %s\n\n",
			fsev.GetUUID(), utils.CGR_SUPPLIERS, fsArray)); err != nil {
			utils.Logger.Info(fmt.Sprintf("<%s> error setting suppliers: %s",
				utils.FreeSWITCHAgent, err.Error()))
			sm.unparkCall(fsev.GetUUID(), connIdx, fsev.GetCallDestNr(utils.MetaDefault), err.Error())
			return
		}
	}

	sm.unparkCall(fsev.GetUUID(), connIdx,
		fsev.GetCallDestNr(utils.MetaDefault), AUTH_OK)
}

func (sm *FSsessions) onChannelAnswer(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.META_NONE { // Do not process this request
		return
	}
	if connIdx >= len(sm.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(sm.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	_, err := sm.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s %s %s\n\n", fsev.GetUUID(),
			utils.CGROriginHost, utils.FirstNonEmpty(sm.cfg.EventSocketConns[connIdx].Alias,
				sm.cfg.EventSocketConns[connIdx].Address)))
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error %s setting channel variabile: %s",
				utils.FreeSWITCHAgent, err.Error(), VarCGROriginHost))
		return
	}
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], sm.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	chanUUID := fsev.GetUUID()
	if missing := fsev.MissingParameter(sm.timezone); missing != "" {
		sm.disconnectSession(connIdx, chanUUID, "",
			utils.NewErrMandatoryIeMissing(missing).Error())
		return
	}
	initSessionArgs := fsev.V1InitSessionArgs()
	initSessionArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID so we can properly disconnect later
	var initReply sessions.V1InitSessionReply
	if err := sm.connMgr.Call(sm.cfg.SessionSConns, sm, utils.SessionSv1InitiateSession,
		initSessionArgs, &initReply); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> could not process answer for event %s, error: %s",
				utils.FreeSWITCHAgent, chanUUID, err.Error()))
		sm.disconnectSession(connIdx, chanUUID, "", err.Error())
		return
	}
}

func (sm *FSsessions) onChannelHangupComplete(fsev FSEvent, connIdx int) {
	if fsev.GetReqType(utils.MetaDefault) == utils.META_NONE { // Do not process this request
		return
	}
	if connIdx >= len(sm.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(sm.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return
	}
	var reply string
	fsev[VarCGROriginHost] = utils.FirstNonEmpty(fsev[VarCGROriginHost], sm.cfg.EventSocketConns[connIdx].Alias) // rewrite the OriginHost variable if it is empty
	if fsev[VarAnswerEpoch] != "0" {                                                                             // call was answered
		terminateSessionArgs := fsev.V1TerminateSessionArgs()
		terminateSessionArgs.CGREvent.Event[FsConnID] = connIdx // Attach the connection ID in case we need to create a session and disconnect it
		if err := sm.connMgr.Call(sm.cfg.SessionSConns, sm, utils.SessionSv1TerminateSession,
			terminateSessionArgs, &reply); err != nil {
			utils.Logger.Err(
				fmt.Sprintf("<%s> Could not terminate session with event %s, error: %s",
					utils.FreeSWITCHAgent, fsev.GetUUID(), err.Error()))
		}
	}
	if sm.cfg.CreateCdr {
		cgrEv, err := fsev.AsCGREvent(sm.timezone)
		if err != nil {
			return
		}
		cgrArgs := cgrEv.ExtractArgs(strings.Index(fsev[VarCGRFlags], utils.MetaDispatchers) != -1, false)
		if err := sm.connMgr.Call(sm.cfg.SessionSConns, sm, utils.SessionSv1ProcessCDR,
			&utils.CGREventWithArgDispatcher{CGREvent: cgrEv, ArgDispatcher: cgrArgs.ArgDispatcher}, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed processing CGREvent: %s,  error: <%s>",
				utils.FreeSWITCHAgent, utils.ToJSON(cgrEv), err.Error()))
		}
	}
}

// Connects to the freeswitch mod_event_socket server and starts
// listening for events.
func (sm *FSsessions) Connect() error {
	eventFilters := map[string][]string{"Call-Direction": {"inbound"}}
	errChan := make(chan error)
	for connIdx, connCfg := range sm.cfg.EventSocketConns {
		fSock, err := fsock.NewFSock(connCfg.Address, connCfg.Password, connCfg.Reconnects,
			sm.createHandlers(), eventFilters, utils.Logger.GetSyslog(), connIdx)
		if err != nil {
			return err
		}
		if !fSock.Connected() {
			return errors.New("Could not connect to FreeSWITCH")
		}
		sm.conns[connIdx] = fSock
		utils.Logger.Info(fmt.Sprintf("<%s> successfully connected to FreeSWITCH at: <%s>", utils.FreeSWITCHAgent, connCfg.Address))
		go func(fsock *fsock.FSock) { // Start reading in own goroutine, return on error
			if err := fsock.ReadEvents(); err != nil {
				errChan <- err
			}
		}(fSock)
		fsSenderPool, err := fsock.NewFSockPool(5, connCfg.Address, connCfg.Password, 1, sm.cfg.MaxWaitConnection,
			make(map[string][]func(string, int)), make(map[string][]string), utils.Logger.GetSyslog(), connIdx)
		if err != nil {
			return fmt.Errorf("Cannot connect FreeSWITCH senders pool, error: %s", err.Error())
		}
		if fsSenderPool == nil {
			return errors.New("Cannot connect FreeSWITCH senders pool.")
		}
		sm.senderPools[connIdx] = fsSenderPool
	}
	err := <-errChan // Will keep the Connect locked until the first error in one of the connections
	return err
}

// fsev.GetCallDestNr(utils.MetaDefault)
// Disconnects a session by sending hangup command to freeswitch
func (sm *FSsessions) disconnectSession(connIdx int, uuid, redirectNr, notify string) error {
	if _, err := sm.conns[connIdx].SendApiCmd(
		fmt.Sprintf("uuid_setvar %s cgr_notify %s\n\n", uuid, notify)); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: %s when attempting to disconnect channelID: %s over connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), uuid, connIdx))
		return err
	}
	if notify == utils.ErrInsufficientCredit.Error() {
		if len(sm.cfg.EmptyBalanceContext) != 0 {
			if _, err := sm.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_transfer %s %s XML %s\n\n",
				uuid, redirectNr, sm.cfg.EmptyBalanceContext)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not transfer the call to empty balance context, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
				return err
			}
			return nil
		}
		if len(sm.cfg.EmptyBalanceAnnFile) != 0 {
			if _, err := sm.conns[connIdx].SendApiCmd(fmt.Sprintf("uuid_broadcast %s playback!manager_request::%s aleg\n\n",
				uuid, sm.cfg.EmptyBalanceAnnFile)); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> Could not send uuid_broadcast to freeswitch, error: <%s>, connIdx: %v",
					utils.FreeSWITCHAgent, err.Error(), connIdx))
				return err
			}
			return nil
		}
	}
	if err := sm.conns[connIdx].SendMsgCmd(uuid,
		map[string]string{"call-command": "hangup", "hangup-cause": "MANAGER_REQUEST"}); err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> Could not send disconect msg to freeswitch, error: <%s>, connIdx: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
		return err
	}
	return nil
}

// Shutdown stops all connected fsock connections
func (sm *FSsessions) Shutdown() (err error) {
	for connIdx, fSock := range sm.conns {
		if !fSock.Connected() {
			utils.Logger.Err(fmt.Sprintf("<%s> Cannot shutdown sessions, fsock not connected for connection index: %v", utils.FreeSWITCHAgent, connIdx))
			continue
		}
		utils.Logger.Info(fmt.Sprintf("<%s> Shutting down all sessions on connection index: %v", utils.FreeSWITCHAgent, connIdx))
		if _, err = fSock.SendApiCmd("hupall MANAGER_REQUEST cgr_reqtype *prepaid"); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on calls shutdown: %s, connection index: %v", utils.FreeSWITCHAgent, err.Error(), connIdx))
		}
		if err = fSock.Disconnect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on disconnect: %s, connection index: %v", utils.FreeSWITCHAgent, err.Error(), connIdx))
		}

	}
	return
}

// rpcclient.ClientConnector interface
func (sm *FSsessions) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(sm, serviceMethod, args, reply)
}

// V1DisconnectSession internal method to disconnect session in FreeSWITCH
func (fsa *FSsessions) V1DisconnectSession(args utils.AttrDisconnectSession, reply *string) (err error) {
	ev := engine.NewMapEvent(args.EventStart)
	channelID := ev.GetStringIgnoreErrors(utils.OriginID)
	connIdx, err := ev.GetTInt64(FsConnID)
	if err != nil {
		utils.Logger.Err(
			fmt.Sprintf("<%s> error: <%s:%s> when attempting to disconnect channelID: <%s>",
				utils.FreeSWITCHAgent, err.Error(), FsConnID, channelID))
		return
	}
	if int(connIdx) >= len(fsa.conns) { // protection against index out of range panic
		err := fmt.Errorf("Index out of range[0,%v): %v ", len(fsa.conns), connIdx)
		utils.Logger.Err(fmt.Sprintf("<%s> %s", utils.FreeSWITCHAgent, err.Error()))
		return err
	}
	if err = fsa.disconnectSession(int(connIdx), channelID,
		utils.FirstNonEmpty(ev.GetStringIgnoreErrors(CALL_DEST_NR), ev.GetStringIgnoreErrors(SIP_REQ_USER)),
		utils.ErrInsufficientCredit.Error()); err != nil {
		return
	}
	*reply = utils.OK
	return
}

func (fsa *FSsessions) V1GetActiveSessionIDs(ignParam string,
	sessionIDs *[]*sessions.SessionID) (err error) {
	var sIDs []*sessions.SessionID
	for connIdx, senderPool := range fsa.senderPools {
		fsConn, err := senderPool.PopFSock()
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on pop FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		activeChanStr, err := fsConn.SendApiCmd("show channels")
		senderPool.PushFSock(fsConn)
		if err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> Error on push FSock: %s, connection index: %v",
				utils.FreeSWITCHAgent, err.Error(), connIdx))
			continue
		}
		aChans := fsock.MapChanData(activeChanStr)
		for _, fsAChan := range aChans {
			sIDs = append(sIDs, &sessions.SessionID{
				OriginHost: fsa.cfg.EventSocketConns[connIdx].Alias,
				OriginID:   fsAChan["uuid"]},
			)
		}
	}
	*sessionIDs = sIDs
	return
}

// Reload recreates the connection buffers
// only used on reload
func (sm *FSsessions) Reload() {
	sm.conns = make([]*fsock.FSock, len(sm.cfg.EventSocketConns))
	sm.senderPools = make([]*fsock.FSockPool, len(sm.cfg.EventSocketConns))
}
