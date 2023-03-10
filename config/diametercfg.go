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

package config

import (
	"strings"

	"github.com/Omnitouch/cgrates/utils"
)

type DiameterAgentCfg struct {
	Enabled           bool   // enables the diameter agent: <true|false>
	ListenNet         string // sctp or tcp
	Listen            string // address where to listen for diameter requests <x.y.z.y:1234>
	DictionariesPath  string
	SessionSConns     []string
	OriginHost        string
	OriginRealm       string
	VendorId          int
	ProductName       string
	ConcurrentReqs    int // limit the maximum number of requests processed
	SyncedConnReqs    bool
	ASRTemplate       string
	Templates         map[string][]*FCTemplate
	RequestProcessors []*RequestProcessor
}

func (da *DiameterAgentCfg) loadFromJsonCfg(jsnCfg *DiameterAgentJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Enabled != nil {
		da.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Listen != nil {
		da.Listen = *jsnCfg.Listen
	}
	if jsnCfg.Listen_net != nil {
		da.ListenNet = *jsnCfg.Listen_net
	}
	if jsnCfg.Dictionaries_path != nil {
		da.DictionariesPath = *jsnCfg.Dictionaries_path
	}
	if jsnCfg.Sessions_conns != nil {
		da.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for idx, attrConn := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if attrConn == utils.MetaInternal {
				da.SessionSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				da.SessionSConns[idx] = attrConn
			}
		}
	}
	if jsnCfg.Origin_host != nil {
		da.OriginHost = *jsnCfg.Origin_host
	}
	if jsnCfg.Origin_realm != nil {
		da.OriginRealm = *jsnCfg.Origin_realm
	}
	if jsnCfg.Vendor_id != nil {
		da.VendorId = *jsnCfg.Vendor_id
	}
	if jsnCfg.Product_name != nil {
		da.ProductName = *jsnCfg.Product_name
	}
	if jsnCfg.Concurrent_requests != nil {
		da.ConcurrentReqs = *jsnCfg.Concurrent_requests
	}
	if jsnCfg.Synced_conn_requests != nil {
		da.SyncedConnReqs = *jsnCfg.Synced_conn_requests
	}
	if jsnCfg.Asr_template != nil {
		da.ASRTemplate = *jsnCfg.Asr_template
	}
	if jsnCfg.Templates != nil {
		if da.Templates == nil {
			da.Templates = make(map[string][]*FCTemplate)
		}
		for k, jsnTpls := range jsnCfg.Templates {
			if da.Templates[k], err = FCTemplatesFromFCTemplatesJsonCfg(jsnTpls, separator); err != nil {
				return
			}
		}
	}
	if jsnCfg.Request_processors != nil {
		for _, reqProcJsn := range *jsnCfg.Request_processors {
			rp := new(RequestProcessor)
			var haveID bool
			for _, rpSet := range da.RequestProcessors {
				if reqProcJsn.ID != nil && rpSet.ID == *reqProcJsn.ID {
					rp = rpSet // Will load data into the one set
					haveID = true
					break
				}
			}
			if err := rp.loadFromJsonCfg(reqProcJsn, separator); err != nil {
				return nil
			}
			if !haveID {
				da.RequestProcessors = append(da.RequestProcessors, rp)
			}
		}
	}
	return nil
}

func (ds *DiameterAgentCfg) AsMapInterface(separator string) map[string]interface{} {
	templates := make(map[string][]map[string]interface{})
	for key, value := range ds.Templates {
		fcTemplate := make([]map[string]interface{}, len(value))
		for i, val := range value {
			fcTemplate[i] = val.AsMapInterface(separator)

		}
		templates[key] = fcTemplate
	}

	requestProcessors := make([]map[string]interface{}, len(ds.RequestProcessors))
	for i, item := range ds.RequestProcessors {
		requestProcessors[i] = item.AsMapInterface(separator)
	}

	sessionSConns := make([]string, len(ds.SessionSConns))
	for i, item := range ds.SessionSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
		if item == buf {
			sessionSConns[i] = strings.ReplaceAll(item, utils.CONCATENATED_KEY_SEP+utils.MetaSessionS, utils.EmptyString)
		} else {
			sessionSConns[i] = item
		}
	}

	return map[string]interface{}{
		utils.EnabledCfg:           ds.Enabled,
		utils.ListenNetCfg:         ds.ListenNet,
		utils.ListenCfg:            ds.Listen,
		utils.DictionariesPathCfg:  ds.DictionariesPath,
		utils.SessionSConnsCfg:     sessionSConns,
		utils.OriginHostCfg:        ds.OriginHost,
		utils.OriginRealmCfg:       ds.OriginRealm,
		utils.VendorIdCfg:          ds.VendorId,
		utils.ProductNameCfg:       ds.ProductName,
		utils.ConcurrentReqsCfg:    ds.ConcurrentReqs,
		utils.SyncedConnReqsCfg:    ds.SyncedConnReqs,
		utils.ASRTemplateCfg:       ds.ASRTemplate,
		utils.TemplatesCfg:         templates,
		utils.RequestProcessorsCfg: requestProcessors,
	}
}
