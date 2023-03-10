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

package v1

import (
	"github.com/Omnitouch/cgrates/utils"
)

// SetTPDispatcherProfile creates a new DispatcherProfile within a tariff plan
func (api *APIerSv1) SetTPDispatcherProfile(attr *utils.TPDispatcherProfile, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPDispatcherProfiles([]*utils.TPDispatcherProfile{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPDispatcherProfile queries specific DispatcherProfile on Tariff plan
func (api *APIerSv1) GetTPDispatcherProfile(attr *utils.TPTntID, reply *utils.TPDispatcherProfile) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rls, err := api.StorDb.GetTPDispatcherProfiles(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

type AttrGetTPDispatcherIds struct {
	TPid string // Tariff plan id
	utils.PaginatorWithSearch
}

// GetTPDispatcherProfileIDs queries dispatcher identities on specific tariff plan.
func (api *APIerSv1) GetTPDispatcherProfileIDs(attrs *AttrGetTPDispatcherIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDispatchers, utils.TPDistinctIds{"tenant", "id"},
		nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPDispatcherProfile removes specific DispatcherProfile on Tariff plan
func (api *APIerSv1) RemoveTPDispatcherProfile(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPDispatchers, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// SetTPDispatcherHost creates a new DispatcherHost within a tariff plan
func (api *APIerSv1) SetTPDispatcherHost(attr *utils.TPDispatcherHost, reply *string) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.SetTPDispatcherHosts([]*utils.TPDispatcherHost{attr}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetTPDispatcherHost queries specific DispatcherHosts on Tariff plan
func (api *APIerSv1) GetTPDispatcherHost(attr *utils.TPTntID, reply *utils.TPDispatcherHost) error {
	if missing := utils.MissingStructFields(attr, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	rls, err := api.StorDb.GetTPDispatcherHosts(attr.TPid, attr.Tenant, attr.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rls[0]
	return nil
}

// GetTPDispatcherHostIDs queries dispatcher host identities on specific tariff plan.
func (api *APIerSv1) GetTPDispatcherHostIDs(attrs *AttrGetTPDispatcherIds, reply *[]string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	ids, err := api.StorDb.GetTpTableIds(attrs.TPid, utils.TBLTPDispatcherHosts, utils.TPDistinctIds{"tenant", "id"},
		nil, &attrs.PaginatorWithSearch)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = ids
	return nil
}

// RemoveTPDispatcherHost removes specific DispatcherHost on Tariff plan
func (api *APIerSv1) RemoveTPDispatcherHost(attrs *utils.TPTntID, reply *string) error {
	if missing := utils.MissingStructFields(attrs, []string{"TPid", "Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := api.StorDb.RemTpData(utils.TBLTPDispatcherHosts, attrs.TPid,
		map[string]string{"tenant": attrs.Tenant, "id": attrs.ID}); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
