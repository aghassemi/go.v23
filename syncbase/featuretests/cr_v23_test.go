// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package featuretests_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"v.io/v23/context"
	"v.io/v23/naming"
	wire "v.io/v23/services/syncbase/nosql"
	"v.io/v23/syncbase"
	"v.io/v23/syncbase/nosql"
	"v.io/v23/verror"
	"v.io/x/ref/lib/v23test"
	"v.io/x/ref/services/syncbase/server/util"
)

// Tests the conflict resolution configuration rules.
// Setup:
// S0 and S1 both have data for keys foo0 to foo9 prepopulated.
// CR Rules are defined as:
//     keys with prefix foo0 -> App based resolution
//     remaining -> default
// S0 and S1 update all rows concurrently causing a conflict for each key.
// Result: foo0 is sent for app based resolution while rest are resolved by
// timestamps.
//
// TODO(jlodhia): Add more rules based on value type and combination of key
// prefix and value type once its implemented.
func TestV23CRRuleConfig(t *testing.T) {
	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 10)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	ok(t, updateData(client0Ctx, "s0", 0, 10, "concurrentUpdate"))
	ok(t, updateData(client1Ctx, "s1", 0, 10, "concurrentUpdate"))

	schemaKeyPrefix := "foo0"
	runWithAppBasedResolver(t, client0Ctx, client1Ctx, schemaKeyPrefix, 2, func() {
		// Re enable sync between the two syncbases and wait for a bit to let the
		// syncbases sync and call conflict resolution.
		ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
		ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

		// Verify that the resolved data looks correct.
		ok(t, waitForValue(client0Ctx, "s0", "foo0", "AppResolvedVal", schemaKeyPrefix))
		ok(t, waitForValue(client0Ctx, "s0", "foo1", "concurrentUpdate"+"s1", schemaKeyPrefix))

		ok(t, waitForValue(client1Ctx, "s1", "foo0", "AppResolvedVal", schemaKeyPrefix))
		ok(t, waitForValue(client1Ctx, "s1", "foo1", "concurrentUpdate"+"s1", schemaKeyPrefix))
	})
}

// Tests the default behavior of conflict resolution, which is last timestamp
// wins, if schema is not specified.
// Setup:
// S0 and S1 both have rows for key foo0. No schema is specified. Both update
// value for foo0 concurrently where S1's write has a newer timestamp.
// Result:
// The value for foo0 after sync settles on what S1 wrote for both syncbases.
func TestV23CRDefault(t *testing.T) {
	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 1)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	ok(t, updateData(client0Ctx, "s0", 0, 1, "concurrentUpdate"))
	time.Sleep(5 * time.Millisecond) // make sure that the clock moves forwared between the two updates.
	ok(t, updateData(client1Ctx, "s1", 0, 1, "concurrentUpdate"))

	// Add new seperate keys to each syncbase so that we can verify if sync
	// has happened between the two syncbases by waiting on the other's key.
	ok(t, populateData(client0Ctx, "s0", "foo", 22, 23))
	ok(t, populateData(client1Ctx, "s1", "foo", 44, 45))

	// Re enable sync between the two syncbases and wait for a bit to let the
	// syncbases sync and call conflict resolution.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

	// Verify that both sides have synced with the other.
	ok(t, waitForValue(client0Ctx, "s0", "foo44", "testkey", "")) // 44 is written by S1
	ok(t, waitForValue(client1Ctx, "s1", "foo22", "testkey", "")) // 22 is written by S0

	// Verify that the resolved data looks correct.
	ok(t, waitForValue(client0Ctx, "s0", "foo0", "concurrentUpdate"+"s1", ""))
	ok(t, waitForValue(client1Ctx, "s1", "foo0", "concurrentUpdate"+"s1", ""))
}

// Tests last timestamp wins for batches under conflict.
// Setup:
// S0 and S1 have prepopulated values for rows foo0 to foo100.
// Conflict resolution type used is LastWins.
// S0 and S1 update all rows in parallel using goroutines so that the actual
// writes are interleaved.
// Result:
// After conflict resolution, final values for all rows within the batch must
// come from either S0 or S1 but not a mixture of the two.
func TestV23CRWithAtomicBatch(t *testing.T) {
	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 100)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	go ok(t, updateDataInBatch(client0Ctx, "s0", 0, 100, "concurrentBatchUpdate", "batchDoneKey1"))
	go ok(t, updateDataInBatch(client1Ctx, "s1", 0, 100, "concurrentBatchUpdate", "batchDoneKey2"))
	time.Sleep(1 * time.Second) // let the above go routine get scheduled.

	// Re enable sync between the two syncbases and wait for a bit to let the
	// syncbases sync and call conflict resolution.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

	// Make sure that the sync has completed by injecting a row on s0 and
	// reading it on s1.
	ok(t, populateData(client0Ctx, "s0", "foo", 200, 201))
	ok(t, waitForValue(client1Ctx, "s1", "foo200", "testkey", ""))
	ok(t, populateData(client1Ctx, "s1", "foo", 400, 401))
	ok(t, waitForValue(client0Ctx, "s0", "foo400", "testkey", ""))

	ok(t, verifyConflictResolvedBatch(client0Ctx, "s0", "foo", 0, 100, "concurrentBatchUpdate"))
	ok(t, verifyConflictResolvedBatch(client1Ctx, "s1", "foo", 0, 100, "concurrentBatchUpdate"))
}

// Tests AppResolves resolution policy by creating conflicts for rows that will
// be resolved by the application. This test covers the following scenerios:
// 1) 5 independent rows under conflict resulting into 5 conflict resolution
//    calls to the app.
// 2) 5 rows written as a single batch on both syncbases resulting into a
//    single conflict for the batch.
func TestV23CRAppResolved(t *testing.T) {
	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 10)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	ok(t, updateData(client0Ctx, "s0", 0, 5, "concurrentUpdate"))
	ok(t, updateData(client1Ctx, "s1", 0, 5, "concurrentUpdate"))

	ok(t, updateDataInBatch(client0Ctx, "s0", 5, 10, "concurrentBatchUpdate", ""))
	ok(t, updateDataInBatch(client1Ctx, "s1", 5, 10, "concurrentBatchUpdate", ""))

	schemaPrefix := "foo"
	keyPrefix := "foo"
	// TODO(jlodhia): change the expected num conflicts from 12 to 6 once
	// sync's cr code handles duplicate resolutions internally.
	runWithAppBasedResolver(t, client0Ctx, client1Ctx, schemaPrefix, 12, func() {
		// Re enable sync between the two syncbases and wait for a bit to let the
		// syncbases sync and call conflict resolution.
		ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
		ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

		// Verify that the resolved data looks correct.
		keyUnderConflict := "foo8" // one of the keys under conflict
		ok(t, waitForValue(client0Ctx, "s0", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client0Ctx, "s0", keyPrefix, schemaPrefix, 0, 5, "AppResolvedVal"))
		ok(t, verifyConflictResolvedData(client0Ctx, "s0", keyPrefix, schemaPrefix, 5, 10, "AppResolvedVal"))

		ok(t, waitForValue(client1Ctx, "s1", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client1Ctx, "s1", keyPrefix, schemaPrefix, 0, 5, "AppResolvedVal"))
		ok(t, verifyConflictResolvedData(client1Ctx, "s1", keyPrefix, schemaPrefix, 5, 10, "AppResolvedVal"))
	})
}

// Tests if a row which was supposed to be resolved based on LastWins policy
// is overridden by AppResolves due to association with a row that has
// AppResolves.
// Setup:
// S0 and S1 have prepopulated values for rows foo0 to foo20.
// Rows with key prefix foo1 are resolved using AppResolves while others are
// LastWins.
// S0 and S1 update all rows concurrently as a single batch where S1's writes
// are newer than S0's.
// Result:
// All rows are resolved via AppResolves.
func TestV23CRAppBasedResolutionOverridesOthers(t *testing.T) {
	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 20)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	ok(t, updateDataInBatch(client0Ctx, "s0", 0, 20, "concurrentBatchUpdate", ""))
	ok(t, updateDataInBatch(client1Ctx, "s1", 0, 20, "concurrentBatchUpdate", ""))

	schemaPrefix := "foo1"
	keyPrefix := "foo"
	// TODO(jlodhia): change the expected num conflicts from 2 to 1 once
	// sync's cr code handles duplicate resolutions internally.
	runWithAppBasedResolver(t, client0Ctx, client1Ctx, schemaPrefix, 2, func() {
		// Re enable sync between the two syncbases and wait for a bit to let the
		// syncbases sync and call conflict resolution.
		ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
		ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

		// Verify that the resolved data looks correct.
		keyUnderConflict := "foo11" // one of the keys under conflict
		ok(t, waitForValue(client0Ctx, "s0", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client0Ctx, "s0", keyPrefix, schemaPrefix, 0, 20, "AppResolvedVal"))

		ok(t, waitForValue(client1Ctx, "s1", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client1Ctx, "s1", keyPrefix, schemaPrefix, 0, 20, "AppResolvedVal"))
	})
}

// Tests if 3 different batches (B1, B2 & B3) that conflict with each other due
// to a subset of rows being common between (B1, B2) and another subset of rows
// common between (B2, B3) results into a single conflict call to the app.
// Setup:
// S0 and S1 have prepopulated values for rows foo0 to foo9.
// Rows with key prefix foo are resolved using AppResolves.
// S0 writes two batches B1{foo0 to foo3}, B3{foo6 to foo9}
// S1 concurrently writes batches B2{foo3 to foo6}
// Result:
// All rows are resolved via AppResolves as a single conflict call.
func TestV23CRMultipleBatchesAsSingleConflict(t *testing.T) {
	// TODO(hpucha): Start running this test once sync handles insertion of
	// local objects by conflict resolution which originally were not under
	// conflict.
	t.Skip()

	sh := v23test.NewShell(t, v23test.Opts{Large: true})
	defer sh.Cleanup()
	client0Ctx, client1Ctx, sgName := setupCRTest(t, sh, 10)

	// Turn off syncing on both s0 and s1 by removing each other from syncgroup ACLs.
	ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0"))
	ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s1"))

	// Since sync is paused, the following updates are concurrent and not
	// racy as long as Put() is sufficiently synchronous.
	// Batch1 has 0, 1, 2, 3 on S0
	ok(t, updateDataInBatch(client0Ctx, "s0", 0, 4, "concurrentBatchUpdate", ""))
	// Batch2 has 6, 7, 8, 9 on S0
	ok(t, updateDataInBatch(client0Ctx, "s0", 6, 10, "concurrentBatchUpdate", ""))
	// Batch3 has 3, 4, 5, 6 on S1
	ok(t, updateDataInBatch(client1Ctx, "s1", 3, 7, "concurrentBatchUpdate", ""))

	schemaPrefix := "foo"
	keyPrefix := "foo"
	// TODO(jlodhia): change the expected num conflicts from 2 to 1 once
	// sync's cr code handles duplicate resolutions internally.
	runWithAppBasedResolver(t, client0Ctx, client1Ctx, schemaPrefix, 2, func() {
		// Re enable sync between the two syncbases and wait for a bit to let the
		// syncbases sync and call conflict resolution.
		ok(t, toggleSync(client0Ctx, "s0", sgName, "root:s0;root:s1"))
		ok(t, toggleSync(client1Ctx, "s1", sgName, "root:s0;root:s1"))

		// Verify that the resolved data looks correct.
		keyUnderConflict := "foo8" // one of the keys under conflict
		ok(t, waitForValue(client0Ctx, "s0", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client0Ctx, "s0", keyPrefix, schemaPrefix, 0, 10, "AppResolvedVal"))

		ok(t, waitForValue(client1Ctx, "s1", keyUnderConflict, "AppResolvedVal", schemaPrefix))
		ok(t, verifyConflictResolvedData(client1Ctx, "s1", keyPrefix, schemaPrefix, 0, 10, "AppResolvedVal"))
	})
}

func setupCRTest(t *testing.T, sh *v23test.Shell, numInitRows int) (client0, client1 *context.T, sgName string) {
	sh.StartRootMountTable()
	sbs := setupSyncbases(t, sh, 2)

	sgName = naming.Join("s0", util.SyncbaseSuffix, "SG1")

	// Create syncgroup and populate data on s0.
	ok(t, createSyncgroup(sbs[0].clientCtx, "s0", sgName, "tb:foo", "", sbBlessings(sbs), nil))
	ok(t, populateData(sbs[0].clientCtx, "s0", "foo", 0, numInitRows))

	// Join syncgroup and verify data on s1.
	ok(t, joinSyncgroup(sbs[1].clientCtx, "s1", sgName))
	ok(t, verifySyncgroupData(sbs[1].clientCtx, "s1", "foo", 0, numInitRows))

	return sbs[0].clientCtx, sbs[1].clientCtx, sgName
}

// TODO(sadovsky): This pattern is not ideal in that it makes the test code
// harder to follow. It would be better to define two helper functions: one
// to do the stuff before fn(), and another to do the stuff after fn(). (Also,
// switching to channel-based signalling should simplify things substantially.)
func runWithAppBasedResolver(t *testing.T, client0Ctx, client1Ctx *context.T, schemaPrefix string, maxCallCount int, fn func()) {
	// Create and hold a conflict resolution connection on s0 and s1 to receive
	// future conflicts. The expected call count is 2 * the number of batches
	// because each batch is being concurrently resolved on s0 and s1 creating new
	// values on each side. Later when the next round of sync happens these new
	// values cause another conflict. Since the conflict resolver does not create
	// new value for a duplicate conflict, no more conflict pingpongs happen.
	// TODO(jlodhia): change the expected num conflicts from 12 to 6 once sync's
	// cr code handles duplicate resolutions internally.
	go func() {
		ok(t, runConflictResolver(client0Ctx, "s0", schemaPrefix, "endKey", maxCallCount))
	}()
	go func() {
		ok(t, runConflictResolver(client1Ctx, "s1", schemaPrefix, "endKey", maxCallCount))
	}()

	time.Sleep(1 * time.Millisecond) // let the above goroutines start up

	fn()

	// endTest signals conflict resolution thread to exit.
	// TODO(sadovsky): Use channels for signaling now that everything's in the
	// same process.
	ok(t, endTest(client0Ctx, "s0", schemaPrefix, "endKey"))
	ok(t, endTest(client1Ctx, "s1", schemaPrefix, "endKey"))

	// wait for conflict resolution thread to exit
	ok(t, waitForSignal(client0Ctx, "s0", schemaPrefix, "endKeyAck"))
	ok(t, waitForSignal(client1Ctx, "s1", schemaPrefix, "endKeyAck"))
}

//////////////////////////////////////////////
// Helpers specific to ConflictResolution

func runConflictResolver(ctx *context.T, syncbaseName, prefix, signalKey string, maxCallCount int) error {
	a := syncbase.NewService(syncbaseName).App("a")
	resolver := &CRImpl{syncbaseName: syncbaseName}
	d := a.NoSQLDatabase("d", makeSchema(prefix, resolver))
	defer d.Close()
	d.EnforceSchema(ctx)

	// Wait till end of test is signalled. The above statement starts a goroutine
	// with a cr connection to the server which needs to stay alive till the life
	// of the test in order to receive conflicts.
	if err := waitSignal(ctx, d, signalKey); err != nil {
		return err
	}

	// Check that OnConflict() was called at most 'maxCallCount' times.
	var onConflictErr error
	if resolver.onConflictCallCount > maxCallCount {
		onConflictErr = fmt.Errorf("Unexpected OnConflict call count. Max: %d, Actual: %d", maxCallCount, resolver.onConflictCallCount)
	}

	// Reply to the test with a signal to notify it that it may end.
	if err := sendSignal(ctx, d, signalKey+"Ack"); err != nil {
		return err
	}

	return onConflictErr
}

func verifyConflictResolvedData(ctx *context.T, syncbaseName, keyPrefix, schemaPrefix string, start, end int, valuePrefix string) error {
	a := syncbase.NewService(syncbaseName).App("a")
	d := a.NoSQLDatabase("d", makeSchema(schemaPrefix, &CRImpl{syncbaseName: syncbaseName}))

	tb := d.Table(testTable)
	for i := start; i < end; i++ {
		var got string
		key := fmt.Sprintf("%s%d", keyPrefix, i)
		r := tb.Row(key)
		if err := r.Get(ctx, &got); err != nil {
			return fmt.Errorf("r.Get() failed: %v", err)
		}
		if got != valuePrefix+key {
			return fmt.Errorf("unexpected value: got %v, want %v", got, valuePrefix)
		}
	}
	return nil
}

func verifyConflictResolvedBatch(ctx *context.T, syncbaseName, keyPrefix string, start, end int, valuePrefix string) error {
	a := syncbase.NewService(syncbaseName).App("a")
	d := a.NoSQLDatabase("d", nil)

	tb := d.Table(testTable)
	var got string

	// get first row
	firstKey := fmt.Sprintf("%s%d", keyPrefix, start)
	r := tb.Row(firstKey)
	if err := r.Get(ctx, &got); err != nil {
		return fmt.Errorf("r.Get() failed: %v\n", err)
	}
	valueServiceStr := strings.TrimSuffix(strings.TrimPrefix(got, valuePrefix), firstKey)

	for i := start; i < end; i++ {
		key := fmt.Sprintf("%s%d", keyPrefix, i)
		r := tb.Row(key)
		if err := r.Get(ctx, &got); err != nil {
			return fmt.Errorf("r.Get() failed: %v\n", err)
		}
		if got != valuePrefix+valueServiceStr+key {
			return fmt.Errorf("unexpected value: got %v, want %v\n", got, valuePrefix+valueServiceStr+key)
		}
	}
	return nil
}

func waitForValue(ctx *context.T, syncbaseName, key, valuePrefix, schemaPrefix string) error {
	var schema *nosql.Schema
	if schemaPrefix != "" {
		schema = makeSchema(schemaPrefix, &CRImpl{syncbaseName: syncbaseName})
	}

	a := syncbase.NewService(syncbaseName).App("a")
	d := a.NoSQLDatabase("d", schema)

	tb := d.Table(testTable)
	r := tb.Row(key)
	want := valuePrefix + key

	// Wait up to 5 seconds for the correct key and value to appear.
	sleepTimeMs, maxAttempts := 50, 100
	var value string
	for i := 0; i < maxAttempts; i++ {
		if err := r.Get(ctx, &value); (err == nil) && (value == want) {
			return nil
		} else if err != nil && verror.ErrorID(err) != verror.ErrNoExist.ID {
			return fmt.Errorf("Syncbase Error while fetching key %v: %v", key, err)
		}
		time.Sleep(time.Duration(sleepTimeMs) * time.Millisecond)
	}
	return fmt.Errorf("Timed out waiting for value %v but found %v after %d milliseconds.", want, value, maxAttempts*sleepTimeMs)
}

func endTest(ctx *context.T, syncbaseName, prefix, signalKey string) error {
	a := syncbase.NewService(syncbaseName).App("a")
	d := a.NoSQLDatabase("d", makeSchema(prefix, &CRImpl{syncbaseName: syncbaseName}))

	// signal end of test so that conflict resolution can clean up its stream.
	return sendSignal(ctx, d, signalKey)
}

func waitForSignal(ctx *context.T, syncbaseName, prefix, signalKey string) error {
	a := syncbase.NewService(syncbaseName).App("a")
	d := a.NoSQLDatabase("d", makeSchema(prefix, &CRImpl{syncbaseName: syncbaseName}))

	// wait for signal.
	return waitSignal(ctx, d, signalKey)
}

func waitSignal(ctx *context.T, d nosql.Database, signalKey string) error {
	tb := d.Table(testTable)
	r := tb.Row(signalKey)

	var end bool
	sleepTimeMs, maxAttempts := 50, 100 // Max wait time of 5 seconds.
	for cnt := 0; cnt < maxAttempts; cnt++ {
		time.Sleep(time.Duration(sleepTimeMs) * time.Millisecond)
		if err := r.Get(ctx, &end); err != nil {
			if verror.ErrorID(err) != verror.ErrNoExist.ID {
				return fmt.Errorf("r.Get() for endkey failed: %v", err)
			}
		}
		if end {
			return nil
		}
	}
	return fmt.Errorf("Timed out waiting for signal %v after %d milliseconds.", signalKey, maxAttempts*sleepTimeMs)
}

////////////////////////////////////////////////////////
// Conflict Resolution related code.

func makeSchema(keyPrefix string, resolver *CRImpl) *nosql.Schema {
	metadata := wire.SchemaMetadata{
		Version: 1,
		Policy: wire.CrPolicy{
			Rules: []wire.CrRule{
				wire.CrRule{
					TableName: testTable,
					KeyPrefix: keyPrefix,
					Resolver:  wire.ResolverTypeAppResolves,
				},
			},
		},
	}
	return &nosql.Schema{
		Metadata: metadata,
		Upgrader: nil,
		Resolver: resolver,
	}
}

// Client conflict resolution impl.
type CRImpl struct {
	syncbaseName        string
	onConflictCallCount int
}

func (ri *CRImpl) OnConflict(ctx *context.T, conflict *nosql.Conflict) nosql.Resolution {
	resolvedPrefix := "AppResolvedVal"
	ri.onConflictCallCount++
	res := nosql.Resolution{ResultSet: map[string]nosql.ResolvedRow{}}
	for rowKey, row := range conflict.WriteSet.ByKey {
		resolvedRow := nosql.ResolvedRow{}
		resolvedRow.Key = row.Key

		// Handle objects that dont have conflict but were pulled in because of
		// other conflicts in the same batch.
		// For ease of testing, this is resolved as new value with prefix
		// "AppResolvedVal".
		if row.LocalValue.State == wire.ValueStateUnknown || row.RemoteValue.State == wire.ValueStateUnknown {
			resolvedRow.Result, _ = nosql.NewValue(ctx, resolvedPrefix+keyPart(rowKey))
			res.ResultSet[row.Key] = resolvedRow
			continue
		}

		var localVal, remoteVal string
		row.LocalValue.Get(&localVal)
		row.RemoteValue.Get(&remoteVal)

		if localVal == remoteVal {
			if row.RemoteValue.WriteTs.After(row.LocalValue.WriteTs) {
				resolvedRow.Result = &row.RemoteValue
			} else {
				resolvedRow.Result = &row.LocalValue
			}
		} else {
			resolvedRow.Result, _ = nosql.NewValue(ctx, resolvedPrefix+keyPart(rowKey))
		}
		res.ResultSet[row.Key] = resolvedRow
	}
	return res
}

func keyPart(rowKey string) string {
	return util.SplitKeyParts(rowKey)[1]
}
