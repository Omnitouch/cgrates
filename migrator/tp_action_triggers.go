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

package migrator

import (
	"fmt"

	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func (m *Migrator) migrateCurrentTPactiontriggers() (err error) {
	tpids, err := m.storDBIn.StorDB().GetTpIds(utils.TBLTPActionTriggers)
	if err != nil {
		return err
	}
	for _, tpid := range tpids {
		ids, err := m.storDBIn.StorDB().GetTpTableIds(tpid, utils.TBLTPActionTriggers, utils.TPDistinctIds{"tag"}, map[string]string{}, nil)
		if err != nil {
			return err
		}
		for _, id := range ids {
			actTrg, err := m.storDBIn.StorDB().GetTPActionTriggers(tpid, id)
			if err != nil {
				return err
			}
			if actTrg != nil {
				if m.dryRun != true {
					if err := m.storDBOut.StorDB().SetTPActionTriggers(actTrg); err != nil {
						return err
					}
					for _, act := range actTrg {
						if err := m.storDBIn.StorDB().RemTpData(
							utils.TBLTPActionTriggers, act.TPid, map[string]string{"tag": act.ID}); err != nil {
							return err
						}
					}
					m.stats[utils.TpActionTriggers] += 1
				}
			}
		}
	}
	return
}

func (m *Migrator) migrateTPactiontriggers() (err error) {
	var vrs engine.Versions
	current := engine.CurrentStorDBVersions()
	vrs, err = m.storDBOut.StorDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.TpActionTriggers] {
	case current[utils.TpActionTriggers]:
		if m.sameStorDB {
			break
		}
		if err := m.migrateCurrentTPactiontriggers(); err != nil {
			return err
		}
	}
	return m.ensureIndexesStorDB(utils.TBLTPActionTriggers)
}
