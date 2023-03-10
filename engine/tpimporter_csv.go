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
	"fmt"
	"io/ioutil"
	"log"

	"github.com/Omnitouch/cgrates/utils"
)

// Import tariff plan from csv into storDb
type TPCSVImporter struct {
	TPid     string     // Load data on this tpid
	StorDb   LoadWriter // StorDb connection handle
	DirPath  string     // Directory path to import from
	Sep      rune       // Separator in the csv file
	Verbose  bool       // If true will print a detailed information instead of silently discarding it
	ImportId string     // Use this to differentiate between imports (eg: when autogenerating fields like RatingProfileID
	csvr     LoadReader
}

// Maps csv file to handler which should process it. Defined like this since tests on 1.0.3 were failing on Travis.
// Change it to func(string) error as soon as Travis updates.
var fileHandlers = map[string]func(*TPCSVImporter, string) error{
	utils.TimingsCsv:            (*TPCSVImporter).importTimings,
	utils.DestinationsCsv:       (*TPCSVImporter).importDestinations,
	utils.RatesCsv:              (*TPCSVImporter).importRates,
	utils.DestinationRatesCsv:   (*TPCSVImporter).importDestinationRates,
	utils.RatingPlansCsv:        (*TPCSVImporter).importRatingPlans,
	utils.RatingProfilesCsv:     (*TPCSVImporter).importRatingProfiles,
	utils.SharedGroupsCsv:       (*TPCSVImporter).importSharedGroups,
	utils.ActionsCsv:            (*TPCSVImporter).importActions,
	utils.ActionPlansCsv:        (*TPCSVImporter).importActionTimings,
	utils.ActionTriggersCsv:     (*TPCSVImporter).importActionTriggers,
	utils.AccountActionsCsv:     (*TPCSVImporter).importAccountActions,
	utils.ResourcesCsv:          (*TPCSVImporter).importResources,
	utils.StatsCsv:              (*TPCSVImporter).importStats,
	utils.ThresholdsCsv:         (*TPCSVImporter).importThresholds,
	utils.FiltersCsv:            (*TPCSVImporter).importFilters,
	utils.SuppliersCsv:          (*TPCSVImporter).importSuppliers,
	utils.AttributesCsv:         (*TPCSVImporter).importAttributeProfiles,
	utils.ChargersCsv:           (*TPCSVImporter).importChargerProfiles,
	utils.DispatcherProfilesCsv: (*TPCSVImporter).importDispatcherProfiles,
	utils.DispatcherHostsCsv:    (*TPCSVImporter).importDispatcherHosts,
}

func (self *TPCSVImporter) Run() error {
	self.csvr = NewFileCSVStorage(self.Sep, self.DirPath, false)
	files, _ := ioutil.ReadDir(self.DirPath)
	var withErrors bool
	for _, f := range files {
		fHandler, hasName := fileHandlers[f.Name()]
		if !hasName {
			continue
		}
		if err := fHandler(self, f.Name()); err != nil {
			withErrors = true
			utils.Logger.Err(fmt.Sprintf("<TPCSVImporter> Importing file: %s, got error: %s", f.Name(), err.Error()))
		}
	}
	if withErrors {
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// Handler importing timings from file, saved row by row to storDb
func (self *TPCSVImporter) importTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPTimings(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPTimings(tps)
}

func (self *TPCSVImporter) importDestinations(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPDestinations(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPDestinations(tps)
}

func (self *TPCSVImporter) importRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPRates(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPRates(tps)
}

func (self *TPCSVImporter) importDestinationRates(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPDestinationRates(self.TPid, "", nil)
	if err != nil {
		return err
	}

	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPDestinationRates(tps)
}

func (self *TPCSVImporter) importRatingPlans(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPRatingPlans(self.TPid, "", nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPRatingPlans(tps)
}

func (self *TPCSVImporter) importRatingProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPRatingProfiles(&utils.TPRatingProfile{TPid: self.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate rating profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}

	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
		tps[i].LoadId = loadId

	}
	return self.StorDb.SetTPRatingProfiles(tps)
}

func (self *TPCSVImporter) importSharedGroups(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPSharedGroups(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPSharedGroups(tps)
}

func (self *TPCSVImporter) importActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPActions(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPActions(tps)
}

func (self *TPCSVImporter) importActionTimings(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPActionPlans(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPActionPlans(tps)
}

func (self *TPCSVImporter) importActionTriggers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPActionTriggers(self.TPid, "")
	if err != nil {
		return err
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
	}

	return self.StorDb.SetTPActionTriggers(tps)
}

func (self *TPCSVImporter) importAccountActions(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	tps, err := self.csvr.GetTPAccountActions(&utils.TPAccountActions{TPid: self.TPid})
	if err != nil {
		return err
	}
	loadId := utils.CSV_LOAD //Autogenerate rating profile id
	if self.ImportId != "" {
		loadId += "_" + self.ImportId
	}
	for i := 0; i < len(tps); i++ {
		tps[i].TPid = self.TPid
		tps[i].LoadId = loadId
	}
	return self.StorDb.SetTPAccountActions(tps)
}

func (self *TPCSVImporter) importResources(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := self.csvr.GetTPResources(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPResources(rls)
}

func (self *TPCSVImporter) importStats(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := self.csvr.GetTPStats(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPStats(sts)
}

func (self *TPCSVImporter) importThresholds(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := self.csvr.GetTPThresholds(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPThresholds(sts)
}

func (self *TPCSVImporter) importFilters(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	sts, err := self.csvr.GetTPFilters(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPFilters(sts)
}

func (self *TPCSVImporter) importSuppliers(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := self.csvr.GetTPSuppliers(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPSuppliers(rls)
}

func (self *TPCSVImporter) importAttributeProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := self.csvr.GetTPAttributes(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPAttributes(rls)
}

func (self *TPCSVImporter) importChargerProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	rls, err := self.csvr.GetTPChargers(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPChargers(rls)
}

func (self *TPCSVImporter) importDispatcherProfiles(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := self.csvr.GetTPDispatcherProfiles(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPDispatcherProfiles(dpps)
}

func (self *TPCSVImporter) importDispatcherHosts(fn string) error {
	if self.Verbose {
		log.Printf("Processing file: <%s> ", fn)
	}
	dpps, err := self.csvr.GetTPDispatcherHosts(self.TPid, "", "")
	if err != nil {
		return err
	}
	return self.StorDb.SetTPDispatcherHosts(dpps)
}
