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
	"testing"
	"time"

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/utils"
)

var (
	matchEV             map[string]interface{}
	dmMatch             *DataManager
	stringIndexedFields *[]string
	prefixIndexedFields *[]string
)

func TestFilterMatchingItemIDsForEvent(t *testing.T) {
	var stringFilter, prefixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dmMatch = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	context := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(attribStringF)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.Field", []string{"profilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF)
	prefix := utils.ConcatenatedKey(config.CgrConfig().GeneralCfg().DefaultTenant, context)
	atrRFI := NewFilterIndexer(dmMatch, utils.AttributeProfilePrefix, prefix)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribStringF), stringFilterID)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribPrefF), prefixFilterID)
	err = atrRFI.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	matchEV = map[string]interface{}{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"Field":          "profile",
	}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}
	matchEV = map[string]interface{}{
		"Field": "profilePrefix",
	}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}

func TestFilterMatchingItemIDsForEvent2(t *testing.T) {
	var stringFilter, prefixFilter, defaultFilter []*FilterRule
	stringFilterID := "stringFilterID"
	prefixFilterID := "prefixFilterID"
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dmMatch = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	context := utils.MetaRating
	x, err := NewFilterRule(utils.MetaString, "~*req.CallCost.Account", []string{"1001"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	stringFilter = append(stringFilter, x)
	attribStringF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "stringFilter",
		Rules:  stringFilter}
	dmMatch.SetFilter(attribStringF)
	x, err = NewFilterRule(utils.MetaPrefix, "~*req.CallCost.Field", []string{"profile"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	prefixFilter = append(prefixFilter, x)
	attribPrefF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "prefFilter",
		Rules:  prefixFilter}
	dmMatch.SetFilter(attribPrefF)
	x, err = NewFilterRule(utils.MetaGreaterOrEqual, "~*req.Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultFilter = append(defaultFilter, x)
	attribDefaultF := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "defaultFilter",
		Rules:  defaultFilter}
	dmMatch.SetFilter(attribDefaultF)
	prefix := utils.ConcatenatedKey(config.CgrConfig().GeneralCfg().DefaultTenant, context)
	atrRFI := NewFilterIndexer(dmMatch, utils.AttributeProfilePrefix, prefix)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribStringF), stringFilterID)
	atrRFI.IndexTPFilter(FilterToTPFilter(attribPrefF), prefixFilterID)
	err = atrRFI.StoreIndexes(true, utils.NonTransactional)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	matchEV = map[string]interface{}{
		utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
		"CallCost":       map[string]interface{}{"Account": 1001},
	}
	aPrflIDs, err := MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has := aPrflIDs[stringFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", stringFilterID, aPrflIDs)
	}
	matchEV = map[string]interface{}{
		"CallCost": map[string]interface{}{"Field": "profilePrefix"},
	}
	aPrflIDs, err = MatchingItemIDsForEvent(matchEV, nil, nil,
		dmMatch, utils.CacheAttributeFilterIndexes, prefix, true, true)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	_, has = aPrflIDs[prefixFilterID]
	if !has {
		t.Errorf("Expecting: %+v, received: %+v", prefixFilterID, aPrflIDs)
	}
}
