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

package utils

import (
	"sort"
	"time"
)

// AccountProfile represents one Account on a Tenant
type AccountProfile struct {
	Tenant             string
	ID                 string // Account identificator, unique within the tenant
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weight             float64
	Opts               map[string]interface{}
	Balances           []*Balance
	ThresholdIDs       []string
}

// Balance represents one Balance inside an Account
type Balance struct {
	ID           string // Balance identificator, unique within an Account
	FilterIDs    []string
	Weight       float64
	Blocker      bool
	Type         string
	Opts         map[string]interface{}
	UsageFactors []*UsageFactor
	Value        float64
}

type UsageFactor struct {
	FilterIDs []string
	Factor    float64
}

func (aP *AccountProfile) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Clone returns a clone of the Account
func (aP *AccountProfile) Clone() (acnt *AccountProfile) {
	return
}

// ActionProfiles is a sortable list of ActionProfiles
type AccountProfiles []*AccountProfile

// Sort is part of sort interface, sort based on Weight
func (aps AccountProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// Balances is a sortable list of Balances
type Balances []*Balance

// Sort is part of sort interface, sort based on Weight
func (blcs Balances) Sort() {
	sort.Slice(blcs, func(i, j int) bool { return blcs[i].Weight > blcs[j].Weight })
}

// AccountProfileWithOpts is used in API calls
type AccountProfileWithOpts struct {
	*AccountProfile
	Opts map[string]interface{}
}

type Account struct {
	Tenant string
	ID     string
}

func (ac *Account) TenantID() string {
	return ConcatenatedKey(ac.Tenant, ac.ID)
}

type AccountWithOpts struct {
	*Account
	Opts map[string]interface{}
}

// ArgsAccountForEvent arguments used for process event
type ArgsAccountForEvent struct {
	*CGREventWithOpts
	AccountIDs []string
}

type ReplyMaxUsage struct {
	AccountID string
	MaxUsage  time.Duration
	Cost      *EventCharges
}