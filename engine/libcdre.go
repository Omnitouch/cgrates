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
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/Omnitouch/cgrates/config"
	"github.com/Omnitouch/cgrates/guardian"
	"github.com/Omnitouch/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var failedPostCache *ltcache.Cache

func init() {
	failedPostCache = ltcache.NewCache(-1, 5*time.Second, false, writeFailedPosts) // configurable  general
}

// SetFailedPostCacheTTL recreates the failed cache
func SetFailedPostCacheTTL(ttl time.Duration) {
	failedPostCache = ltcache.NewCache(-1, ttl, false, writeFailedPosts)
}

func writeFailedPosts(itmID string, value interface{}) {
	expEv, canConvert := value.(*ExportEvents)
	if !canConvert {
		return
	}
	filePath := path.Join(config.CgrConfig().GeneralCfg().FailedPostsDir, expEv.FileName())
	if err := expEv.WriteToFile(filePath); err != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> Failed to write file <%s> because <%s>",
			utils.CDRs, filePath, err))
	}
	return
}

func addFailedPost(expPath, format, module string, ev interface{}) {
	key := utils.ConcatenatedKey(expPath, format, module)
	var failedPost *ExportEvents
	if x, ok := failedPostCache.Get(key); ok {
		if x != nil {
			failedPost = x.(*ExportEvents)
		}
	}
	if failedPost == nil {
		failedPost = &ExportEvents{
			Path:   expPath,
			Format: format,
			module: module,
		}
	}
	failedPost.AddEvent(ev)
	failedPostCache.Set(key, failedPost, nil)
}

// NewExportEventsFromFile returns ExportEvents from the file
// used only on replay failed post
func NewExportEventsFromFile(filePath string) (expEv *ExportEvents, err error) {
	var fileContent []byte
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		if fileContent, err = ioutil.ReadFile(filePath); err != nil {
			return 0, err
		}
		return 0, os.Remove(filePath)
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	if err != nil {
		return
	}
	dec := gob.NewDecoder(bytes.NewBuffer(fileContent))
	// unmarshall it
	expEv = new(ExportEvents)
	err = dec.Decode(&expEv)
	return
}

// ExportEvents used to save the failed post to file
type ExportEvents struct {
	lk     sync.RWMutex
	Path   string
	Format string
	Events []interface{}
	module string
}

// FileName returns the file name it should use for saving the failed events
func (expEv *ExportEvents) FileName() string {
	return expEv.module + utils.HandlerArgSep + utils.UUIDSha1Prefix() + utils.GOBSuffix
}

// SetModule sets the module for this event
func (expEv *ExportEvents) SetModule(mod string) {
	expEv.module = mod
}

// WriteToFile writes the events to file
func (expEv *ExportEvents) WriteToFile(filePath string) (err error) {
	_, err = guardian.Guardian.Guard(func() (interface{}, error) {
		fileOut, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		encd := gob.NewEncoder(fileOut)
		err = encd.Encode(expEv)
		fileOut.Close()
		return nil, err
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.FileLockPrefix+filePath)
	return
}

// AddEvent adds one event
func (expEv *ExportEvents) AddEvent(ev interface{}) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *ExportEvents) ReplayFailedPosts(attempts int) (failedEvents *ExportEvents, err error) {
	failedEvents = &ExportEvents{
		Path:   expEv.Path,
		Format: expEv.Format,
	}
	switch expEv.Format {
	case utils.MetaHTTPjsonCDR, utils.MetaHTTPjsonMap, utils.MetaHTTPjson, utils.MetaHTTPPost:
		var pstr *HTTPPoster
		pstr, err = NewHTTPPoster(config.CgrConfig().GeneralCfg().HttpSkipTlsVerify,
			config.CgrConfig().GeneralCfg().ReplyTimeout, expEv.Path,
			utils.PosterTransportContentTypes[expEv.Format],
			config.CgrConfig().GeneralCfg().PosterAttempts)
		if err != nil {
			return expEv, err
		}
		for _, ev := range expEv.Events {
			err = pstr.Post(ev, utils.EmptyString)
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	case utils.MetaAMQPjsonCDR, utils.MetaAMQPjsonMap:
		for _, ev := range expEv.Events {
			err = PostersCache.PostAMQP(expEv.Path, attempts, ev.([]byte))
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	case utils.MetaAMQPV1jsonMap:
		for _, ev := range expEv.Events {
			err = PostersCache.PostAMQPv1(expEv.Path, attempts, ev.([]byte))
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	case utils.MetaSQSjsonMap:
		for _, ev := range expEv.Events {
			err = PostersCache.PostSQS(expEv.Path, attempts, ev.([]byte))
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	case utils.MetaKafkajsonMap:
		for _, ev := range expEv.Events {
			err = PostersCache.PostKafka(expEv.Path, attempts, ev.([]byte), utils.UUIDSha1Prefix())
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	case utils.MetaS3jsonMap:
		for _, ev := range expEv.Events {
			err = PostersCache.PostS3(expEv.Path, attempts, ev.([]byte), utils.UUIDSha1Prefix())
			if err != nil {
				failedEvents.AddEvent(ev)
			}
		}
	}
	if len(failedEvents.Events) > 0 {
		err = utils.ErrPartiallyExecuted
	} else {
		failedEvents = nil
	}
	return
}
