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

package ers

import (
	"fmt"

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

type EventReader interface {
	Config() *config.EventReaderCfg // return it's configuration
	Serve() error                   // subscribe the reader on the path
}

// NewEventReader instantiates the event reader based on configuration at index
func NewEventReader(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	switch cfg.ERsCfg().Readers[cfgIdx].Type {
	default:
		err = fmt.Errorf("unsupported reader type: <%s>", cfg.ERsCfg().Readers[cfgIdx].Type)
	case utils.MetaFileCSV:
		return NewCSVFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaPartialCSV:
		return NewPartialCSVFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileXML:
		return NewXMLFileER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFileFWV:
		return NewFWVFileERER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaKafkajsonMap:
		return NewKafkaER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaSQL:
		return NewSQLEventReader(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	case utils.MetaFlatstore:
		return NewFlatstoreER(cfg, cfgIdx, rdrEvents, rdrErr, fltrS, rdrExit)
	}
	return
}
