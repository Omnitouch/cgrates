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

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/engine"
	"github.com/Omnitouch/cgrates/utils"
)

type v1Attribute struct {
	FieldName  string
	Initial    string
	Substitute string
	Append     bool
}

type v1AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval          // Activation interval
	Attributes         map[string]map[string]*v1Attribute // map[FieldName][InitialValue]*Attribute
	Weight             float64
}

func (m *Migrator) migrateCurrentAttributeProfile() (err error) {
	var ids []string
	ids, err = m.dmIN.DataManager().DataDB().GetKeysForPrefix(utils.AttributeProfilePrefix)
	if err != nil {
		return err
	}
	for _, id := range ids {
		tntID := strings.SplitN(strings.TrimPrefix(id, utils.AttributeProfilePrefix), utils.InInFieldSep, 2)
		if len(tntID) < 2 {
			return fmt.Errorf("Invalid key <%s> when migrating attributes", id)
		}
		attrPrf, err := m.dmIN.DataManager().GetAttributeProfile(tntID[0], tntID[1], false, false, utils.NonTransactional)
		if err != nil {
			return err
		}
		if attrPrf == nil || m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		if err := m.dmIN.DataManager().RemoveAttributeProfile(tntID[0],
			tntID[1], utils.NonTransactional, false); err != nil {
			return err
		}
		m.stats[utils.Attributes]++
	}
	return
}

func (m *Migrator) migrateV1Attributes() (err error) {
	var v1Attr *v1AttributeProfile
	for {
		v1Attr, err = m.dmIN.getV1AttributeProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v1Attr == nil {
			continue
		}
		attrPrf, err := v1Attr.AsAttributeProfile()
		if err != nil {
			return err
		}
		if m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().DataDB().SetAttributeProfileDrv(attrPrf); err != nil {
			return err
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		m.stats[utils.Attributes]++
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.Attributes: engine.CurrentDataDBVersions()[utils.Attributes]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateV2Attributes() (err error) {
	var v2Attr *v2AttributeProfile
	for {
		v2Attr, err = m.dmIN.getV2AttributeProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v2Attr == nil {
			continue
		}
		attrPrf, err := v2Attr.AsAttributeProfile()
		if err != nil {
			return err
		}
		if m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		m.stats[utils.Attributes]++
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.Attributes: engine.CurrentDataDBVersions()[utils.Attributes]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateV3Attributes() (err error) {
	var v3Attr *v3AttributeProfile
	for {
		v3Attr, err = m.dmIN.getV3AttributeProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v3Attr == nil {
			continue
		}
		attrPrf, err := v3Attr.AsAttributeProfile()
		if err != nil {
			return err
		}
		if m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		m.stats[utils.Attributes]++
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.Attributes: engine.CurrentDataDBVersions()[utils.Attributes]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateV4Attributes() (err error) {
	var v4Attr *v4AttributeProfile
	for {
		v4Attr, err = m.dmIN.getV4AttributeProfile()
		if err != nil && err != utils.ErrNoMoreData {
			return err
		}
		if err == utils.ErrNoMoreData {
			break
		}
		if v4Attr == nil {
			continue
		}
		attrPrf, err := v4Attr.AsAttributeProfile()
		if err != nil {
			return err
		}
		if m.dryRun {
			continue
		}
		if err := m.dmOut.DataManager().SetAttributeProfile(attrPrf, true); err != nil {
			return err
		}
		m.stats[utils.Attributes]++
	}
	if m.dryRun {
		return
	}
	// All done, update version wtih current one
	vrs := engine.Versions{utils.Attributes: engine.CurrentDataDBVersions()[utils.Attributes]}
	if err = m.dmOut.DataManager().DataDB().SetVersions(vrs, false); err != nil {
		return utils.NewCGRError(utils.Migrator,
			utils.ServerErrorCaps,
			err.Error(),
			fmt.Sprintf("error: <%s> when updating Thresholds version into dataDB", err.Error()))
	}
	return
}

func (m *Migrator) migrateAttributeProfile() (err error) {
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
			"version number is not defined for ActionTriggers model")
	}
	switch vrs[utils.Attributes] {
	case current[utils.Attributes]:
		if m.sameDataDB {
			break
		}
		if err = m.migrateCurrentAttributeProfile(); err != nil {
			return err
		}
	case 1:
		if err = m.migrateV1Attributes(); err != nil {
			return err
		}
	case 2:
		if err = m.migrateV2Attributes(); err != nil {
			return err
		}
	case 3:
		if err = m.migrateV3Attributes(); err != nil {
			return err
		}
	case 4:
		if err = m.migrateV4Attributes(); err != nil {
			return err
		}
	}
	return m.ensureIndexesDataDB(engine.ColAttr)
}

func (v1AttrPrf v1AttributeProfile) AsAttributeProfile() (attrPrf *engine.AttributeProfile, err error) {
	attrPrf = &engine.AttributeProfile{
		Tenant:             v1AttrPrf.Tenant,
		ID:                 v1AttrPrf.ID,
		Contexts:           v1AttrPrf.Contexts,
		FilterIDs:          v1AttrPrf.FilterIDs,
		Weight:             v1AttrPrf.Weight,
		ActivationInterval: v1AttrPrf.ActivationInterval,
	}
	for _, mp := range v1AttrPrf.Attributes {
		for _, attr := range mp {
			filterIDs := make([]string, 0)
			//append false translate to  if FieldName exist do stuff
			if attr.Append == false {
				filterIDs = append(filterIDs, utils.MetaExists+":"+attr.FieldName+":")
			}
			//Initial not *any translate to if value of fieldName = initial do stuff
			if attr.Initial != utils.META_ANY {
				filterIDs = append(filterIDs, utils.MetaString+":"+attr.FieldName+":"+attr.Initial)
			}
			sbstPrsr, err := config.NewRSRParsers(attr.Substitute, true, config.CgrConfig().GeneralCfg().RSRSep)
			if err != nil {
				return nil, err
			}
			attrPrf.Attributes = append(attrPrf.Attributes, &engine.Attribute{
				FilterIDs: filterIDs,
				Path:      utils.MetaReq + utils.NestingSep + attr.FieldName,
				Value:     sbstPrsr,
				Type:      utils.MetaVariable,
			})
		}
	}
	return
}

type v2Attribute struct {
	FieldName  string
	Initial    interface{}
	Substitute config.RSRParsers
	Append     bool
}

type v2AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Attributes         []*v2Attribute
	Weight             float64
}

func (v2AttrPrf v2AttributeProfile) AsAttributeProfile() (attrPrf *engine.AttributeProfile, err error) {
	attrPrf = &engine.AttributeProfile{
		Tenant:             v2AttrPrf.Tenant,
		ID:                 v2AttrPrf.ID,
		Contexts:           v2AttrPrf.Contexts,
		FilterIDs:          v2AttrPrf.FilterIDs,
		Weight:             v2AttrPrf.Weight,
		ActivationInterval: v2AttrPrf.ActivationInterval,
	}
	for _, attr := range v2AttrPrf.Attributes {
		filterIDs := make([]string, 0)
		//append false translate to  if FieldName exist do stuff
		if !attr.Append {
			filterIDs = append(filterIDs, utils.MetaExists+":"+attr.FieldName+":")
		}
		//Initial not *any translate to if value of fieldName = initial do stuff
		if attr.Initial != nil { // if is nil we take it as default: utils.META_ANY
			if initStr, canCast := attr.Initial.(string); !canCast {
				return nil, fmt.Errorf("can't cast initial value to string for AttributeProfile with ID:%s", attrPrf.ID)
			} else if initStr != utils.META_ANY {
				filterIDs = append(filterIDs, utils.MetaString+":"+attr.FieldName+":"+initStr)
			}
		}
		attrPrf.Attributes = append(attrPrf.Attributes, &engine.Attribute{
			FilterIDs: filterIDs,
			Path:      utils.MetaReq + utils.NestingSep + attr.FieldName,
			Value:     attr.Substitute,
			Type:      utils.MetaVariable,
		})
	}
	return
}

type v3Attribute struct {
	FilterIDs  []string
	FieldName  string
	Substitute config.RSRParsers
}

type v3AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Attributes         []*v3Attribute
	Weight             float64
}

func (v3AttrPrf v3AttributeProfile) AsAttributeProfile() (attrPrf *engine.AttributeProfile, err error) {
	attrPrf = &engine.AttributeProfile{
		Tenant:             v3AttrPrf.Tenant,
		ID:                 v3AttrPrf.ID,
		Contexts:           v3AttrPrf.Contexts,
		FilterIDs:          v3AttrPrf.FilterIDs,
		Weight:             v3AttrPrf.Weight,
		ActivationInterval: v3AttrPrf.ActivationInterval,
	}
	for _, attr := range v3AttrPrf.Attributes {
		attrPrf.Attributes = append(attrPrf.Attributes, &engine.Attribute{
			FilterIDs: attr.FilterIDs,
			Path:      utils.MetaReq + utils.NestingSep + attr.FieldName,
			Value:     attr.Substitute,
			Type:      utils.MetaVariable, //default value for type
		})
	}
	return
}

type v4Attribute struct {
	FilterIDs []string
	FieldName string
	Type      string
	Value     config.RSRParsers
}

type v4AttributeProfile struct {
	Tenant             string
	ID                 string
	Contexts           []string // bind this AttributeProfile to multiple contexts
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Attributes         []*v4Attribute
	Blocker            bool // blocker flag to stop processing on multiple runs
	Weight             float64
}

func (v4AttrPrf v4AttributeProfile) AsAttributeProfile() (attrPrf *engine.AttributeProfile, err error) {
	attrPrf = &engine.AttributeProfile{
		Tenant:             v4AttrPrf.Tenant,
		ID:                 v4AttrPrf.ID,
		Contexts:           v4AttrPrf.Contexts,
		FilterIDs:          v4AttrPrf.FilterIDs,
		Weight:             v4AttrPrf.Weight,
		ActivationInterval: v4AttrPrf.ActivationInterval,
	}
	for _, attr := range v4AttrPrf.Attributes {
		val := attr.Value.GetRule()
		rsrVal := attr.Value
		if strings.HasPrefix(val, utils.DynamicDataPrefix) {
			val = val[1:] // remove the DynamicDataPrefix
			val = utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + val
			rsrVal, err = config.NewRSRParsers(val, true, config.CgrConfig().GeneralCfg().RSRSep)
			if err != nil {
				return nil, err
			}
		}

		attrPrf.Attributes = append(attrPrf.Attributes, &engine.Attribute{
			FilterIDs: attr.FilterIDs,
			Path:      utils.MetaReq + utils.NestingSep + attr.FieldName,
			Value:     rsrVal,
			Type:      attr.Type,
		})
	}
	return
}
