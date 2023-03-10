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
	"strings"

	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

func (m *Migrator) migrateCurrentSupplierProfile() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.SupplierProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.SupplierProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating supplier profiles", id)
		}
		splp, err := m.dmIN.DataManager().GetSupplierProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if splp == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetSupplierProfile(splp, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveSupplierProfile(tntID[0], tntID[1], utils.NonTransactional, true); err != nil {
			return err
		}
		m.stats[utils.Suppliers] += 1
	}
	return
}

func (m *Migrator) migrateSupplierProfiles() (err error) {
	var vrs engine.Versions
	current := engine.CurrentDataDBVersions()
	vrs, err = m.dmIN.DataManager().DataDB().GetVersions("")
	if err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when querying oldDataDB for versions", err.Error()))
	} else if len(vrs) == 0 {
		return utils.NewCGRError(utils.Migrator,
			utils.MandatoryIEMissingCaps,
			utils.UndefinedVersion,
			"version number is not defined for SupplierProfiles model")
	}
	switch vrs[utils.Suppliers] {
	case current[utils.Suppliers]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentSupplierProfile(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColSpp)
}
