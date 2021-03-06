// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tabletserver

import (
	"fmt"
	"html/template"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/youtube/vitess/go/sqldb"
	"github.com/youtube/vitess/go/vt/dbconfigs"
	"github.com/youtube/vitess/go/vt/mysqlctl"
)

type fakeCallInfo struct {
	remoteAddr string
	username   string
	text       string
	html       string
}

func (fci *fakeCallInfo) RemoteAddr() string {
	return fci.remoteAddr
}

func (fci *fakeCallInfo) Username() string {
	return fci.username
}

func (fci *fakeCallInfo) Text() string {
	return fci.text
}

func (fci *fakeCallInfo) HTML() template.HTML {
	return template.HTML(fci.html)
}

type testUtils struct{}

func newTestUtils() *testUtils {
	return &testUtils{}
}

func (util *testUtils) checkEqual(t *testing.T, expected interface{}, result interface{}) {
	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("expect to get: %v, but got: %v", expected, result)
	}
}

func (util *testUtils) checkTabletErrorWithRecover(t *testing.T, tabletErrType int, tabletErrStr string) {
	err := recover()
	if err == nil {
		t.Fatalf("should get error")
	}
	util.checkTabletError(t, err, tabletErrType, tabletErrStr)
}

func (util *testUtils) checkTabletError(t *testing.T, err interface{}, tabletErrType int, tabletErrStr string) {
	tabletError, ok := err.(*TabletError)
	if !ok {
		t.Fatalf("should return a TabletError, but got err: %v", err)
	}
	if tabletError.ErrorType != tabletErrType {
		t.Fatalf("should return a TabletError with error type: %s", util.getTabletErrorString(tabletErrType))
	}
	if !strings.Contains(tabletError.Error(), tabletErrStr) {
		t.Fatalf("expect the tablet error should contain string: '%s', but it does not. Got tablet error: '%s'", tabletErrStr, tabletError.Error())
	}
}

func (util *testUtils) getTabletErrorString(tabletErrorType int) string {
	switch tabletErrorType {
	case ErrFail:
		return "ErrFail"
	case ErrRetry:
		return "ErrRetry"
	case ErrFatal:
		return "ErrFatal"
	case ErrTxPoolFull:
		return "ErrTxPoolFull"
	case ErrNotInTx:
		return "ErrNotInTx"
	}
	return ""
}

func (util *testUtils) newMysqld(dbconfigs *dbconfigs.DBConfigs) *mysqlctl.Mysqld {
	randID := rand.Int63()
	return mysqlctl.NewMysqld(
		fmt.Sprintf("Dba_%d", randID),
		fmt.Sprintf("App_%d", randID),
		mysqlctl.NewMycnf(0, 6802),
		&dbconfigs.Dba,
		&dbconfigs.App.ConnParams,
		&dbconfigs.Repl,
	)
}

func (util *testUtils) newDBConfigs() dbconfigs.DBConfigs {
	appDBConfig := dbconfigs.DBConfig{
		ConnParams:        sqldb.ConnParams{},
		Keyspace:          "test_keyspace",
		Shard:             "0",
		EnableRowcache:    false,
		EnableInvalidator: false,
	}
	return dbconfigs.DBConfigs{
		App: appDBConfig,
	}
}

func (util *testUtils) newQueryServiceConfig() Config {
	randID := rand.Int63()
	config := DefaultQsConfig
	config.StatsPrefix = fmt.Sprintf("Stats-%d-", randID)
	config.DebugURLPrefix = fmt.Sprintf("/debug-%d-", randID)
	config.RowCache.StatsPrefix = fmt.Sprintf("Stats-%d-", randID)
	config.PoolNamePrefix = fmt.Sprintf("Pool-%d-", randID)
	config.StrictMode = true
	config.RowCache.Binary = "ls"
	config.RowCache.Connections = 100
	return config
}

func newTestSchemaInfo(
	queryCacheSize int,
	reloadTime time.Duration,
	idleTimeout time.Duration,
	enablePublishStats bool) *SchemaInfo {
	randID := rand.Int63()
	return NewSchemaInfo(
		queryCacheSize,
		fmt.Sprintf("TestSchemaInfo-%d-", randID),
		map[string]string{
			debugQueryPlansKey: fmt.Sprintf("/debug/query_plans_%d", randID),
			debugQueryStatsKey: fmt.Sprintf("/debug/query_stats_%d", randID),
			debugTableStatsKey: fmt.Sprintf("/debug/table_stats_%d", randID),
			debugSchemaKey:     fmt.Sprintf("/debug/schema_%d", randID),
		},
		reloadTime,
		idleTimeout,
		enablePublishStats,
	)
}
