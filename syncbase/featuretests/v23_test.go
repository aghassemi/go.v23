// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated via go generate.
// DO NOT UPDATE MANUALLY

package featuretests_test

import (
	"os"
	"testing"

	"v.io/x/ref/test/modules"
	"v.io/x/ref/test/v23tests"
)

func TestMain(m *testing.M) {
	modules.DispatchAndExitIfChild()
	cleanup := v23tests.UseSharedBinDir()
	r := m.Run()
	cleanup()
	os.Exit(r)
}

func TestV23BlobWholeTransfer(t *testing.T) {
	v23tests.RunTest(t, V23TestBlobWholeTransfer)
}

func TestV23SyncbasedPutGet(t *testing.T) {
	v23tests.RunTest(t, V23TestSyncbasedPutGet)
}

func TestV23CRRuleConfig(t *testing.T) {
	v23tests.RunTest(t, V23TestCRRuleConfig)
}

func TestV23CRDefault(t *testing.T) {
	v23tests.RunTest(t, V23TestCRDefault)
}

func TestV23CRWithAtomicBatch(t *testing.T) {
	v23tests.RunTest(t, V23TestCRWithAtomicBatch)
}

func TestV23CRAppResolved(t *testing.T) {
	v23tests.RunTest(t, V23TestCRAppResolved)
}

func TestV23CRAppBasedResolutionOverridesOthers(t *testing.T) {
	v23tests.RunTest(t, V23TestCRAppBasedResolutionOverridesOthers)
}

func TestV23CRMultipleBatchesAsSingleConflict(t *testing.T) {
	v23tests.RunTest(t, V23TestCRMultipleBatchesAsSingleConflict)
}

func TestV23DeviceManager(t *testing.T) {
	v23tests.RunTest(t, V23TestDeviceManager)
}

func TestV23RestartabilityHierarchy(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityHierarchy)
}

func TestV23RestartabilityCrash(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityCrash)
}

func TestV23RestartabilityQuiescent(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityQuiescent)
}

func TestV23RestartabilityReadOnlyBatch(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityReadOnlyBatch)
}

func TestV23RestartabilityReadWriteBatch(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityReadWriteBatch)
}

func TestV23RestartabilityWatch(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityWatch)
}

func TestV23RestartabilityServiceDBCorruption(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityServiceDBCorruption)
}

func TestV23RestartabilityAppDBCorruption(t *testing.T) {
	v23tests.RunTest(t, V23TestRestartabilityAppDBCorruption)
}

func TestV23SyncgroupRendezvousOnline(t *testing.T) {
	v23tests.RunTest(t, V23TestSyncgroupRendezvousOnline)
}

func TestV23SyncgroupRendezvousOnlineCloud(t *testing.T) {
	v23tests.RunTest(t, V23TestSyncgroupRendezvousOnlineCloud)
}

func TestV23SyncgroupNeighborhoodOnly(t *testing.T) {
	v23tests.RunTest(t, V23TestSyncgroupNeighborhoodOnly)
}

func TestV23SyncgroupPreknownStaggered(t *testing.T) {
	v23tests.RunTest(t, V23TestSyncgroupPreknownStaggered)
}

func TestV23VClockMovesForward(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockMovesForward)
}

func TestV23VClockSystemClockUpdate(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockSystemClockUpdate)
}

func TestV23VClockSystemClockFrequency(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockSystemClockFrequency)
}

func TestV23VClockNtpUpdate(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockNtpUpdate)
}

func TestV23VClockNtpSkewAfterReboot(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockNtpSkewAfterReboot)
}

func TestV23VClockNtpFrequency(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockNtpFrequency)
}

func TestV23VClockSyncBasic(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockSyncBasic)
}

func TestV23VClockSyncWithLocalNtp(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockSyncWithLocalNtp)
}

func TestV23VClockSyncWithReboots(t *testing.T) {
	v23tests.RunTest(t, V23TestVClockSyncWithReboots)
}
