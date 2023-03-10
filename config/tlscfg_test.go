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
	"reflect"
	"testing"

	"github.com/Omnitouch/cgrates/utils"
)

func TestTlsCfgloadFromJsonCfg(t *testing.T) {
	var tlscfg, expected TlsCfg
	if err := tlscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tlscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, tlscfg)
	}
	if err := tlscfg.loadFromJsonCfg(new(TlsJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tlscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, tlscfg)
	}
	cfgJSONStr := `	{
	"tls":{
		"server_certificate" : "path/To/Server/Cert",			
		"server_key":"path/To/Server/Key",					
		"client_certificate" : "path/To/Client/Cert",			
		"client_key":"path/To/Client/Key",					
		"ca_certificate":"path/To/CA/Cert",							
		"server_name":"TestServerName",
		"server_policy":3,					
		},
	}`
	expected = TlsCfg{
		ServerCerificate: "path/To/Server/Cert",
		ServerKey:        "path/To/Server/Key",
		CaCertificate:    "path/To/CA/Cert",
		ClientCerificate: "path/To/Client/Cert",
		ClientKey:        "path/To/Client/Key",
		ServerName:       "TestServerName",
		ServerPolicy:     3,
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, tlscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, tlscfg)
	}
}
func TestTlsCfgAsMapInterface(t *testing.T) {
	var tlscfg TlsCfg
	cfgJSONStr := `	{
	"tls": {
		"server_certificate" : "",			
		"server_key":"",					
		"client_certificate" : "",			
		"client_key":"",					
		"ca_certificate":"",				
		"server_policy":4,					
		"server_name":"",
	},
}`
	eMap := map[string]interface{}{
		"server_certificate": "",
		"server_key":         "",
		"client_certificate": "",
		"client_key":         "",
		"ca_certificate":     "",
		"server_policy":      4,
		"server_name":        "",
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if rcv := tlscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v,\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `	{
	"tls":{
		"server_certificate" : "path/To/Server/Cert",			
		"server_key":"path/To/Server/Key",					
		"client_certificate" : "path/To/Client/Cert",			
		"client_key":"path/To/Client/Key",					
		"ca_certificate":"path/To/CA/Cert",							
		"server_name":"TestServerName",
		"server_policy":3,					
	},
}`
	eMap = map[string]interface{}{
		"server_certificate": "path/To/Server/Cert",
		"server_key":         "path/To/Server/Key",
		"client_certificate": "path/To/Client/Cert",
		"client_key":         "path/To/Client/Key",
		"ca_certificate":     "path/To/CA/Cert",
		"server_name":        "TestServerName",
		"server_policy":      3,
	}

	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsntlsCfg, err := jsnCfg.TlsCfgJson(); err != nil {
		t.Error(err)
	} else if err = tlscfg.loadFromJsonCfg(jsntlsCfg); err != nil {
		t.Error(err)
	} else if rcv := tlscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v,\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
