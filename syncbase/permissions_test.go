// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncbase_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/security"
	"v.io/v23/security/access"
	wire "v.io/v23/services/syncbase"
	"v.io/v23/syncbase"
	"v.io/v23/syncbase/util"
	"v.io/v23/verror"
	vsecurity "v.io/x/ref/lib/security"
	_ "v.io/x/ref/runtime/factories/roaming"
	tu "v.io/x/ref/services/syncbase/testutil"
)

type serviceTest struct {
	f func(ctx *context.T, s syncbase.Service) error
}

type databaseTest struct {
	f func(ctx *context.T, d syncbase.Database) error
}

type batchDatabaseTest struct {
	f func(ctx *context.T, b syncbase.BatchDatabase) error
}

type collectionTest struct {
	f func(ctx *context.T, c syncbase.Collection) error
}

type rowTest struct {
	f func(ctx *context.T, r syncbase.Row) error
}

type securitySpecTest struct {
	layer   interface{}
	name    string
	pattern string
}

type securitySpecTestGroup struct {
	layer    interface{}
	name     string
	patterns []string
	mutating bool
}

// TODO(rogulenko): Add more test groups.
var securitySpecTestGroups = []securitySpecTestGroup{
	// Service tests.
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			_, err := wire.ServiceClient(s.FullName()).DevModeGetTime(ctx)
			return err
		}},
		name:     "service.DevModeGetTime",
		patterns: []string{"A___"},
	},
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			return wire.ServiceClient(s.FullName()).DevModeUpdateVClock(ctx, wire.DevModeUpdateVClockOpts{})
		}},
		name:     "service.DevModeUpdateVClock",
		patterns: []string{"A___"},
		mutating: true,
	},
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			_, _, err := s.GetPermissions(ctx)
			return err
		}},
		name:     "service.GetPermissions",
		patterns: []string{"A___"},
	},
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			return s.SetPermissions(ctx, tu.DefaultPerms(access.AllTypicalTags(), "root"), "")
		}},
		name:     "service.SetPermissions",
		patterns: []string{"A___"},
		mutating: true,
	},
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			_, err := s.ListDatabases(ctx)
			return err
		}},
		name:     "service.ListDatabases",
		patterns: []string{"R___"},
	},

	// Database tests.
	{
		layer: serviceTest{f: func(ctx *context.T, s syncbase.Service) error {
			return s.DatabaseForId(wire.Id{"root", "dNew"}, nil).Create(ctx, tu.DefaultPerms(wire.AllDatabaseTags, "root"))
		}},
		name:     "database.Create",
		patterns: []string{"W___"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return d.Destroy(ctx)
		}},
		name:     "database.Destroy",
		patterns: []string{"XA__", "A___"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := d.Exists(ctx)
			return err
		}},
		name:     "database.Exists",
		patterns: []string{"XX__", "XR__", "XW__", "XA__", "R___", "W___"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, _, err := d.GetPermissions(ctx)
			return err
		}},
		name:     "database.GetPermissions",
		patterns: []string{"XA__"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return d.SetPermissions(ctx, tu.DefaultPerms(wire.AllDatabaseTags, "root"), "")
		}},
		name:     "database.SetPermissions",
		patterns: []string{"XA__"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := d.ListCollections(ctx)
			return err
		}},
		name:     "database.ListCollections",
		patterns: []string{"XR__"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			ws := d.Watch(ctx, nil, []wire.CollectionRowPattern{util.RowPrefixPattern(wire.Id{"u", "c"}, "")})
			ws.Advance()
			return ws.Err()
		}},
		name:     "database.Watch",
		patterns: []string{"XR__"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := d.BeginBatch(ctx, wire.BatchOptions{})
			return err
		}},
		name:     "database.BeginBatch",
		patterns: []string{"XX__", "XR__", "XW__", "XA__"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := d.GetResumeMarker(ctx)
			return err
		}},
		name:     "database.GetResumeMarker",
		patterns: []string{"XX__", "XR__", "XW__", "XA__"},
	},
	// TODO(ivanpi): Test Exec.
	// TODO(ivanpi): Test RPC-only methods such as Glob.

	// Batch database tests.
	{
		layer: batchDatabaseTest{f: func(ctx *context.T, b syncbase.BatchDatabase) error {
			return b.Commit(ctx)
		}},
		name:     "database.Commit",
		patterns: []string{"XX__", "XR__", "XW__", "XA__"},
		mutating: true,
	},
	{
		layer: batchDatabaseTest{f: func(ctx *context.T, b syncbase.BatchDatabase) error {
			return b.Abort(ctx)
		}},
		name:     "database.Abort",
		patterns: []string{"XX__", "XR__", "XW__", "XA__"},
	},

	// Conflict resolver tests.
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := wire.DatabaseClient(d.FullName()).GetSchemaMetadata(ctx)
			if verror.ErrorID(err) == verror.ErrNoExist.ID {
				return nil
			}
			return err
		}},
		name:     "database.GetSchemaMetadata",
		patterns: []string{"XR__"},
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return wire.DatabaseClient(d.FullName()).SetSchemaMetadata(ctx, wire.SchemaMetadata{})
		}},
		name:     "database.SetSchemaMetadata",
		patterns: []string{"XA__"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return wire.DatabaseClient(d.FullName()).PauseSync(ctx)
		}},
		name:     "database.PauseSync",
		patterns: []string{"XA__"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return wire.DatabaseClient(d.FullName()).ResumeSync(ctx)
		}},
		name:     "database.ResumeSync",
		patterns: []string{"XA__"},
		mutating: true,
	},
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			s, err := wire.DatabaseClient(d.FullName()).StartConflictResolver(ctx)
			if err != nil {
				return err
			}
			go s.RecvStream().Advance()
			<-time.After(1 * time.Second)
			return s.RecvStream().Err()
		}},
		name:     "database.StartConflictResolver",
		patterns: []string{"XA__"},
		mutating: true,
	},

	// Blob manager tests.
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			_, err := d.CreateBlob(ctx)
			return err
		}},
		name:     "database.CreateBlob",
		patterns: []string{"XW__"},
		mutating: true,
	},
	// TODO(ivanpi): Test other blob RPCs.

	// Syncgroup manager tests.
	// TODO(ivanpi): Add.

	// Collection tests.
	{
		layer: databaseTest{f: func(ctx *context.T, d syncbase.Database) error {
			return d.CollectionForId(wire.Id{"root", "cNew"}).Create(ctx, tu.DefaultPerms(wire.AllCollectionTags, "root"))
		}},
		name:     "collection.Create",
		patterns: []string{"XW__"},
		mutating: true,
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			return c.Destroy(ctx)
		}},
		name:     "collection.Destroy",
		patterns: []string{"XXA_", "XA__"},
		mutating: true,
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			_, err := c.Exists(ctx)
			return err
		}},
		name:     "collection.Exists",
		patterns: []string{"XXR_", "XXW_", "XXA_", "XR__", "XW__"},
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			_, err := c.GetPermissions(ctx)
			return err
		}},
		name:     "collection.GetPermissions",
		patterns: []string{"XXA_"},
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			return c.SetPermissions(ctx, tu.DefaultPerms(wire.AllCollectionTags, "root"))
		}},
		name:     "collection.SetPermissions",
		patterns: []string{"XXA_"},
		mutating: true,
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			ss := c.Scan(ctx, syncbase.Prefix(""))
			ss.Advance()
			return ss.Err()
		}},
		name:     "collection.Scan",
		patterns: []string{"XXR_"},
	},
	{
		layer: collectionTest{f: func(ctx *context.T, c syncbase.Collection) error {
			return c.DeleteRange(ctx, syncbase.Prefix(""))
		}},
		name:     "collection.DeleteRange",
		patterns: []string{"XXW_"},
		mutating: true,
	},
	// TODO(ivanpi): Test RPC-only methods such as Glob.

	// Row tests.
	{
		layer: rowTest{f: func(ctx *context.T, r syncbase.Row) error {
			_, err := r.Exists(ctx)
			return err
		}},
		name:     "row.Exists",
		patterns: []string{"XXR_", "XXW_"},
	},
	{
		layer: rowTest{f: func(ctx *context.T, r syncbase.Row) error {
			var value string
			return r.Get(ctx, &value)
		}},
		name:     "row.Get",
		patterns: []string{"XXR_"},
	},
	{
		layer: rowTest{f: func(ctx *context.T, r syncbase.Row) error {
			value := "NCC-1701-D"
			return r.Put(ctx, &value)
		}},
		name:     "row.Put",
		patterns: []string{"XXW_"},
		mutating: true,
	},
	{
		layer: rowTest{f: func(ctx *context.T, r syncbase.Row) error {
			return r.Delete(ctx)
		}},
		name:     "row.Delete",
		patterns: []string{"XXW_"},
		mutating: true,
	},
}

// TestSecuritySpec tests the Syncbase ACL specification. It runs a set of test
// groups.
//
// Each test group describes a Syncbase public method and a list of security
// patterns that should allow a client to call the method.
// A method might be service.GetPermissions and the pattern for it is "A___".
// A security pattern is a string of 4 bytes where each byte is an '_' or
// one of X (resolve), R (read), W (write), A (admin).
// The character index stands for:
// 0 - service ACL
// 1 - database ACL
// 2 - collection ACL
// 3 - syncgroup ACL
// A pattern defines a set of per-ACL permissions required for a client to call
// the method.
//
// For each test group we generate two sets of tests: allowed tests and denied
// tests.
// Allowed tests are generated by splitting the list of security patterns
// of the test group into one pattern per test.
// Denied tests are generated the following way: we take each allowed pattern
// and generate all patterns that differ in one character, preserving all '_'
// (excluding patterns that are a subset of an allowed pattern). This way we
// make sure that every character in a security pattern is important.
// Allowed tests are split in two categories: mutating tests and static tests.
// Mutating tests change the state of Syncbase, static don't.
// Denied tests are always static.
//
// Then we pack all tests into runs, where each run has only one type of tests
// (allowed or denied) and at most one mutating test. For each run we rebuild
// the following Syncbase structure:
// service s; database {a,d}; collection c; row prefix.
//
// For each test inside a run we generate 5 huge ACLs with a record for each
// of the test + one admin record.
func TestSecuritySpec(t *testing.T) {
	deniedTests, staticAllowedTests, mutatingAllowedTests := prepareTests()
	runTests(t, false, deniedTests...)
	runTests(t, true, staticAllowedTests...)
	for _, test := range mutatingAllowedTests {
		runTests(t, true, test)
	}
}

func prepareTests() (deniedTests, staticAllowedTests, mutatingAllowedTests []securitySpecTest) {
	for _, testGroup := range securitySpecTestGroups {
		// Collect allowed patterns.
		allowedPatterns := map[string]bool{}
		for _, pattern := range testGroup.patterns {
			allowedPatterns[pattern] = true
		}
		// Fill in allowed test slices.
		for pattern := range allowedPatterns {
			test := securitySpecTest{
				layer:   testGroup.layer,
				name:    testGroup.name,
				pattern: pattern,
			}
			if testGroup.mutating {
				mutatingAllowedTests = append(mutatingAllowedTests, test)
			} else {
				staticAllowedTests = append(staticAllowedTests, test)
			}
		}

		// Collect denied patterns.
		deniedPatterns := map[string]bool{}
		for _, pattern := range testGroup.patterns {
			for i := 0; i < len(pattern); i++ {
				if pattern[i] != '_' {
					patternBytes := []byte(pattern)
					for _, c := range []byte{'X', 'R', 'W', 'A'} {
						patternBytes[i] = c
						if !isSubsetOfAny(string(patternBytes), allowedPatterns) {
							deniedPatterns[string(patternBytes)] = true
						}
					}
				}
			}
		}
		// Fill in denied test slice.
		for pattern := range deniedPatterns {
			test := securitySpecTest{
				layer:   testGroup.layer,
				name:    testGroup.name,
				pattern: pattern,
			}
			deniedTests = append(deniedTests, test)
		}
	}
	return
}

func isSubsetOfAny(needle string, haystack map[string]bool) bool {
	for p := range haystack {
		isSubset := true
		for i := 0; i < len(p); i++ {
			if p[i] != '_' && p[i] != needle[i] {
				isSubset = false
				break
			}
		}
		if isSubset {
			// All perms in needle correspond to the same perm or '_' in p.
			return true
		}
	}
	return false
}

func runTests(t *testing.T, expectSuccess bool, tests ...securitySpecTest) {
	// Create permissions.
	servicePerms := makePerms("root:admin", 0, access.AllTypicalTags(), tests...)
	databasePerms := makePerms("root:admin", 1, wire.AllDatabaseTags, tests...)
	collectionPerms := makePerms("root:admin", 2, wire.AllCollectionTags, tests...)
	//sgPerms := makePerms("root:admin", 3, wire.AllSyncgroupTags, tests...)

	// Create service/database/collection/row with permissions above.
	ctx, adminCtx, sName, rootp, cleanup := tu.SetupOrDieCustom("admin", "server", nil)
	defer cleanup()
	s := syncbase.NewService(sName)
	if err := s.SetPermissions(adminCtx, servicePerms, ""); err != nil {
		tu.Fatalf(t, "s.SetPermissions failed: %v", err)
	}
	d := s.DatabaseForId(wire.Id{"root", "d"}, nil)
	if err := d.Create(adminCtx, databasePerms); err != nil {
		tu.Fatalf(t, "d.Create failed: %v", err)
	}
	c := d.CollectionForId(wire.Id{"root:admin", "c"})
	if err := c.Create(adminCtx, collectionPerms); err != nil {
		tu.Fatalf(t, "c.Create failed: %v", err)
	}
	r := c.Row("prefix")
	r.Put(adminCtx, "value")

	// Verify tests.
	for i, test := range tests {
		clientCtx := tu.NewCtx(ctx, rootp, fmt.Sprintf("client%d", i))
		var err error
		switch layer := test.layer.(type) {
		case serviceTest:
			err = layer.f(clientCtx, s)
		case databaseTest:
			err = layer.f(clientCtx, d)
		case batchDatabaseTest:
			b, bErr := d.BeginBatch(adminCtx, wire.BatchOptions{})
			if bErr != nil {
				tu.Fatalf(t, "d.BeginBatch failed: %v", bErr)
			}
			err = layer.f(clientCtx, b)
		case collectionTest:
			err = layer.f(clientCtx, c)
		case rowTest:
			err = layer.f(clientCtx, r)
		default:
			tu.Fatal(t, "unknown test type")
		}
		if expectSuccess && err != nil {
			tu.Fatalf(t, "test %v failed with non-nil error: %v", test, err)
		} else if !expectSuccess && verror.ErrorID(err) != verror.ErrNoAccess.ID && verror.ErrorID(err) != verror.ErrNoExistOrNoAccess.ID {
			// TODO(ivanpi): Allow specifying which error is expected. Also test on nonexistent layers.
			tu.Fatalf(t, "test %v didn't fail with ErrNoAccess or ErrNoExistOrNoAccess error: %v", test, err)
		}
	}
}

// makePerms creates a perms object with permissions for each test.
// For each test we add a blessing pattern "root:clientXX" with a tag
// corresponding test.pattern[index].
func makePerms(admin string, index int, allowTags []access.Tag, tests ...securitySpecTest) access.Permissions {
	perms := tu.DefaultPerms(allowTags, admin)
	tagsMap := map[byte]string{
		'X': "Resolve",
		'R': "Read",
		'W': "Write",
		'A': "Admin",
	}
	for i, t := range tests {
		if tag, ok := tagsMap[t.pattern[index]]; ok {
			perms.Add(security.BlessingPattern(fmt.Sprintf("root:client%d", i)), tag)
		}
	}
	return util.FilterTags(perms, allowTags...)
}

var (
	inferBlessingsOk = []struct {
		exts     []string
		wantApp  string
		wantUser string
	}{
		{
			[]string{"o:angrybirds:alice"},
			"root:o:angrybirds", "root:o:angrybirds:alice",
		},
		{
			[]string{"foo", "o:angrybirds:alice", "x:baz"},
			"root:o:angrybirds", "root:o:angrybirds:alice",
		},
		{
			[]string{"foo", "u:alice", "o:angrybirds:alice:device:phone", "x:baz"},
			"root:o:angrybirds", "root:o:angrybirds:alice",
		},
		{
			[]string{"foo", "u:bob", "o:todos:dave:friend:alice", "u:carol", "x:baz"},
			"root:o:todos", "root:o:todos:dave",
		},
		{
			[]string{"foo", "u:bob", "o:todos:dave:friend:alice", "u:carol", "o:todos:dave:device:phone", "x:baz"},
			"root:o:todos", "root:o:todos:dave",
		},
		{
			[]string{"u:bob"},
			"...", "root:u:bob",
		},
		{
			[]string{"foo", "u:bob", "x:baz"},
			"...", "root:u:bob",
		},
		{
			[]string{"foo", "u:bob:angrybirds", "u:bob:todos:phone", "x:baz"},
			"...", "root:u:bob",
		},
	}

	inferBlessingsFail = [][]string{
		{},
		{"foo", "x:baz"},
		{"foo", "u:bob", "o:todos:dave:friend:alice", "o:angrybirds:dave", "o:todos:dave:device:phone", "x:baz"},
		{"foo", "u:bob", "o:todos:dave:friend:alice", "o:todos:fred"},
		{"foo", "u:bob:angrybirds", "u:bob:todos:phone", "u:carol", "u:dave", "x:baz"},
	}

	dummyPerms = access.Permissions{}.Add("root:u:nobody", string(access.Admin))
)

// TestIdBlessingInfer tests that Database/Collection/Syncgroup getter variants
// taking a context and name correctly infer the Id blessing from the context,
// failing when ambiguous.
func TestIdBlessingInfer(t *testing.T) {
	anchorPerms := access.Permissions{}.Add("root:u:admin", string(access.Admin)).Add("root", string(access.Resolve), string(access.Read), string(access.Write))
	rootCtx, adminCtx, sName, rootp, cleanup := tu.SetupOrDieCustom("u:admin", "server", anchorPerms)
	defer cleanup()
	s := syncbase.NewService(sName)
	rd := s.Database(adminCtx, "anchor", nil)
	if err := rd.Create(adminCtx, anchorPerms); err != nil {
		t.Fatalf("rd.Create() failed: %v", err)
	}
	rc := rd.Collection(adminCtx, "anchor")
	if err := rc.Create(adminCtx, util.FilterTags(anchorPerms, wire.AllCollectionTags...)); err != nil {
		t.Fatalf("rc.Create() failed: %v", err)
	}
	sgSpec := wire.SyncgroupSpec{
		Description: "dummy syncgroup",
		Collections: []wire.Id{rc.Id()},
		Perms:       dummyPerms,
	}
	sgInfo := wire.SyncgroupMemberInfo{
		SyncPriority: 42,
	}

	for i, test := range inferBlessingsOk {
		ctx := forkContextMultiPattern(rootCtx, rootp, test.exts)
		nameSuffix := fmt.Sprintf("ok_%02d", i)

		d := s.Database(ctx, "db_"+nameSuffix, nil)
		if got, want := d.Id(), (wire.Id{Blessing: test.wantApp, Name: "db_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred database id %v, got %v", test.exts, want, got)
		}
		if err := d.Create(ctx, dummyPerms); err != nil {
			t.Errorf("blessings %v: d.Create() failed: %v", test.exts, err)
		}

		c := rd.Collection(ctx, "cx_"+nameSuffix)
		if got, want := c.Id(), (wire.Id{Blessing: test.wantUser, Name: "cx_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred collection id %v, got %v", test.exts, want, got)
		}
		if err := c.Create(ctx, dummyPerms); err != nil {
			t.Errorf("blessings %v: c.Create() failed: %v", test.exts, err)
		}

		sg := rd.Syncgroup(ctx, "sg_"+nameSuffix)
		if got, want := sg.Id(), (wire.Id{Blessing: test.wantUser, Name: "sg_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred syncgroup id %v, got %v", test.exts, want, got)
		}
		if err := sg.Create(ctx, sgSpec, sgInfo); err != nil {
			t.Errorf("blessings %v: sg.Create() failed: %v", test.exts, err)
		}
	}

	for i, exts := range inferBlessingsFail {
		ctx := forkContextMultiPattern(rootCtx, rootp, exts)
		nameSuffix := fmt.Sprintf("notok_%02d", i)

		d := s.Database(ctx, "db_"+nameSuffix, nil)
		if got, want := d.Id(), (wire.Id{Blessing: "$", Name: "db_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred database id %v, got %v", exts, want, got)
		}
		if err := d.Create(ctx, dummyPerms); verror.ErrorID(err) != wire.ErrInvalidName.ID {
			t.Errorf("blessings %v: d.Create() should have failed with ErrInvalidName, got: %v", exts, err)
		}

		c := rd.Collection(ctx, "cx_"+nameSuffix)
		if got, want := c.Id(), (wire.Id{Blessing: "$", Name: "cx_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred collection id %v, got %v", exts, want, got)
		}
		if err := c.Create(ctx, dummyPerms); verror.ErrorID(err) != wire.ErrInvalidName.ID {
			t.Errorf("blessings %v: c.Create() should have failed with ErrInvalidName, got: %v", exts, err)
		}

		sg := rd.Syncgroup(ctx, "sg_"+nameSuffix)
		if got, want := sg.Id(), (wire.Id{Blessing: "$", Name: "sg_" + nameSuffix}); got != want {
			t.Errorf("blessings %v: expected inferred syncgroup id %v, got %v", exts, want, got)
		}
		if err := sg.Create(ctx, sgSpec, sgInfo); verror.ErrorID(err) != wire.ErrInvalidName.ID {
			t.Errorf("blessings %v: sg.Create() should have failed with ErrInvalidName, got: %v", exts, err)
		}
	}
}

// TestIdBlessingInfer tests that Database/Collection Create() calls correctly
// infer default perms from the context when nil perms are passed in, failing
// when ambiguous.
func TestCreatePermsInfer(t *testing.T) {
	anchorPerms := access.Permissions{}.Add("root:u:admin", string(access.Admin)).Add("root", string(access.Resolve), string(access.Read), string(access.Write))
	rootCtx, adminCtx, sName, rootp, cleanup := tu.SetupOrDieCustom("u:admin", "server", anchorPerms)
	defer cleanup()
	s := syncbase.NewService(sName)
	rd := s.Database(adminCtx, "anchor", nil)
	if err := rd.Create(adminCtx, anchorPerms); err != nil {
		t.Fatalf("rd.Create() failed: %v", err)
	}

	for i, test := range inferBlessingsOk {
		ctx := forkContextMultiPattern(rootCtx, rootp, test.exts)
		nameSuffix := fmt.Sprintf("ok_%02d", i)

		d := s.DatabaseForId(wire.Id{Blessing: "root", Name: "db_" + nameSuffix}, nil)
		if err := d.Create(ctx, nil); err != nil {
			t.Errorf("blessings %v: d.Create() failed: %v", test.exts, err)
		}
		if got, _, err := d.GetPermissions(ctx); err != nil {
			t.Fatalf("d.GetPermissions() failed: %v", err)
		} else if want := tu.DefaultPerms(wire.AllDatabaseTags, test.wantUser); !reflect.DeepEqual(got, want) {
			t.Errorf("blessings %v: expected inferred database perms %v, got %v", test.exts, want, got)
		}

		c := rd.CollectionForId(wire.Id{Blessing: "root", Name: "cx_" + nameSuffix})
		if err := c.Create(ctx, nil); err != nil {
			t.Errorf("blessings %v: c.Create() failed: %v", test.exts, err)
		}
		if got, err := c.GetPermissions(ctx); err != nil {
			t.Fatalf("c.GetPermissions() failed: %v", err)
		} else if want := tu.DefaultPerms(wire.AllCollectionTags, test.wantUser); !reflect.DeepEqual(got, want) {
			t.Errorf("blessings %v: expected inferred collection perms %v, got %v", test.exts, want, got)
		}
	}

	for i, exts := range inferBlessingsFail {
		if len(exts) == 0 {
			continue
		}
		ctx := forkContextMultiPattern(rootCtx, rootp, exts)
		nameSuffix := fmt.Sprintf("notok_%02d", i)

		d := s.DatabaseForId(wire.Id{Blessing: "root", Name: "db_" + nameSuffix}, nil)
		if err := d.Create(ctx, nil); verror.ErrorID(err) != wire.ErrInferDefaultPermsFailed.ID {
			t.Errorf("blessings %v: d.Create() should have failed with ErrInferDefaultPermsFailed, got: %v", exts, err)
		}
		if err := d.Create(ctx, dummyPerms); err != nil {
			t.Errorf("blessings %v: d.Create() with explicit perms failed: %v", exts, err)
		}

		c := rd.CollectionForId(wire.Id{Blessing: "root", Name: "cx_" + nameSuffix})
		if err := c.Create(ctx, nil); verror.ErrorID(err) != wire.ErrInferDefaultPermsFailed.ID {
			t.Errorf("blessings %v: c.Create() should have failed with ErrInferDefaultPermsFailed, got: %v", exts, err)
		}
		if err := c.Create(ctx, dummyPerms); err != nil {
			t.Errorf("blessings %v: c.Create() with explicit perms failed: %v", exts, err)
		}
	}
}

func forkContextMultiPattern(rootCtx *context.T, rootp security.Principal, extensions []string) *context.T {
	p, _ := vsecurity.NewPrincipal()
	db, _ := rootp.BlessingStore().Default()
	security.AddToRoots(p, db)
	bs := make([]security.Blessings, 0, len(extensions))
	for _, ext := range extensions {
		b, _ := rootp.Bless(p.PublicKey(), db, ext, security.UnconstrainedUse())
		bs = append(bs, b)
	}
	b, _ := security.UnionOfBlessings(bs...)
	p.BlessingStore().SetDefault(b)
	p.BlessingStore().Set(b, "...")
	ctx, _ := v23.WithPrincipal(rootCtx, p)
	return ctx
}

// TestCreateIdBlessingEnforce tests that only clients matching the blessing
// pattern in the Id are allowed to create a Database/Collection/Syncgroup.
// This requirement is waived for admin clients for Database, but not Collection
// or Syncgroup creation.
func TestCreateIdBlessingEnforce(t *testing.T) {
	anchorPerms := access.Permissions{}.Add("root:u:admin", string(access.Admin)).Add("root", string(access.Resolve), string(access.Read), string(access.Write))
	rootCtx, adminCtx, sName, rootp, cleanup := tu.SetupOrDieCustom("u:admin", "server", anchorPerms)
	defer cleanup()
	clientCtx := tu.NewCtx(rootCtx, rootp, "u:client")
	s := syncbase.NewService(sName)
	rd := s.Database(adminCtx, "anchor", nil)
	if err := rd.Create(adminCtx, anchorPerms); err != nil {
		t.Fatalf("rd.Create() failed: %v", err)
	}
	rc := rd.Collection(adminCtx, "anchor")
	if err := rc.Create(adminCtx, util.FilterTags(anchorPerms, wire.AllCollectionTags...)); err != nil {
		t.Fatalf("rc.Create() failed: %v", err)
	}
	sgSpec := wire.SyncgroupSpec{
		Description: "dummy syncgroup",
		Collections: []wire.Id{rc.Id()},
		Perms:       dummyPerms,
	}
	sgInfo := wire.SyncgroupMemberInfo{
		SyncPriority: 42,
	}

	patternsOk := []string{"...", "root", "root:u", "root:u:client", "root:u:client:$"}
	patternsFail := []string{"root:$", "root:u:$", "root:u:admin", "root:u:client:phone", "foobar"}

	for i, pattern := range patternsOk {
		nameSuffix := fmt.Sprintf("ok_%02d", i)

		d := s.DatabaseForId(wire.Id{Blessing: pattern, Name: "db_" + nameSuffix}, nil)
		if err := d.Create(clientCtx, dummyPerms); err != nil {
			t.Errorf("blessing %q: d.Create() failed: %v", pattern, err)
		}

		c := rd.CollectionForId(wire.Id{Blessing: pattern, Name: "cx_" + nameSuffix})
		if err := c.Create(clientCtx, dummyPerms); err != nil {
			t.Errorf("blessing %q: c.Create() failed: %v", pattern, err)
		}

		sg := rd.SyncgroupForId(wire.Id{Blessing: pattern, Name: "sg_" + nameSuffix})
		if err := sg.Create(clientCtx, sgSpec, sgInfo); err != nil {
			t.Errorf("blessing %q: sg.Create() failed: %v", pattern, err)
		}
	}

	for i, pattern := range patternsFail {
		nameSuffix := fmt.Sprintf("notok_%02d", i)

		d := s.DatabaseForId(wire.Id{Blessing: pattern, Name: "db_" + nameSuffix}, nil)
		if err := d.Create(clientCtx, dummyPerms); verror.ErrorID(err) != wire.ErrUnauthorizedCreateId.ID {
			t.Errorf("blessing %q: d.Create() should have failed with ErrUnauthorizedCreateId, got: %v", pattern, err)
		}

		c := rd.CollectionForId(wire.Id{Blessing: pattern, Name: "cx_" + nameSuffix})
		if err := c.Create(clientCtx, dummyPerms); verror.ErrorID(err) != wire.ErrUnauthorizedCreateId.ID {
			t.Errorf("blessing %q: c.Create() should have failed with ErrUnauthorizedCreateId, got: %v", pattern, err)
		}

		sg := rd.SyncgroupForId(wire.Id{Blessing: pattern, Name: "sg_" + nameSuffix})
		if err := sg.Create(clientCtx, sgSpec, sgInfo); verror.ErrorID(err) != wire.ErrUnauthorizedCreateId.ID {
			t.Errorf("blessing %q: sg.Create() should have failed with ErrUnauthorizedCreateId, got: %v", pattern, err)
		}
	}

	// Give full admin permissions to the client.
	clientAdminPerms := anchorPerms.Copy().Add("root:u:client", string(access.Admin))
	if err := s.SetPermissions(adminCtx, clientAdminPerms, ""); err != nil {
		t.Fatalf("s.SetPermissions() failed: %v", err)
	}
	if err := rd.SetPermissions(adminCtx, clientAdminPerms, ""); err != nil {
		t.Fatalf("rd.SetPermissions() failed: %v", err)
	}
	if err := rc.SetPermissions(adminCtx, util.FilterTags(clientAdminPerms, wire.AllCollectionTags...)); err != nil {
		t.Fatalf("rc.SetPermissions() failed: %v", err)
	}

	// Admin permissions override the Id blessing enforcement for databases, but
	// not for collections or syncgroups.
	for i, pattern := range patternsFail {
		nameSuffix := fmt.Sprintf("admin_%02d", i)

		d := s.DatabaseForId(wire.Id{Blessing: pattern, Name: "db_" + nameSuffix}, nil)
		if err := d.Create(clientCtx, dummyPerms); err != nil {
			t.Errorf("blessing %q: admin d.Create() failed: %v", pattern, err)
		}

		c := rd.CollectionForId(wire.Id{Blessing: pattern, Name: "cx_" + nameSuffix})
		if err := c.Create(clientCtx, dummyPerms); verror.ErrorID(err) != wire.ErrUnauthorizedCreateId.ID {
			t.Errorf("blessing %q: admin c.Create() should have failed with ErrUnauthorizedCreateId, got: %v", pattern, err)
		}

		sg := rd.SyncgroupForId(wire.Id{Blessing: pattern, Name: "sg_" + nameSuffix})
		if err := sg.Create(clientCtx, sgSpec, sgInfo); verror.ErrorID(err) != wire.ErrUnauthorizedCreateId.ID {
			t.Errorf("blessing %q: admin sg.Create() should have failed with ErrUnauthorizedCreateId, got: %v", pattern, err)
		}
	}
}

// TestPermsValidation tests that all operations that create or modify
// permissions properly validate the new permissions.
func TestPermsValidation(t *testing.T) {
	ctx, sName, cleanup := tu.SetupOrDie(nil)
	defer cleanup()
	s := syncbase.NewService(sName)
	rd := tu.CreateDatabase(t, ctx, s, "anchor", nil)
	rc := tu.CreateCollection(t, ctx, rd, "anchor")
	rsg := tu.CreateSyncgroup(t, ctx, rd, rc, "anchor", "anchor syncgroup")

	permsEmpty := access.Permissions{}
	permsNoAdmin := access.Permissions{}.Add("root:o:app:client", string(access.Read))
	permsXRWAD := access.Permissions{}.Add("root:o:app:client", access.TagStrings(access.AllTypicalTags()...)...)
	permsXRWA := access.Permissions{}.Add("root:o:app:client", access.TagStrings(wire.AllDatabaseTags...)...)
	permsRWA := access.Permissions{}.Add("root:o:app:client", access.TagStrings(wire.AllCollectionTags...)...)
	permsRA := access.Permissions{}.Add("root:o:app:client", access.TagStrings(wire.AllSyncgroupTags...)...)
	permsAdminOnly := access.Permissions{}.Add("root:o:app:client", string(access.Admin))
	permsComplex := permsRA.Copy().Add("root:u:client", string(access.Read)).Blacklist("root:o:app:client:phone", string(access.Admin))

	testPermsValidationOp(t, "d.Create()",
		[]access.Permissions{nil /* infer */, permsAdminOnly, permsComplex, permsRA, permsRWA, permsXRWA},
		[]access.Permissions{permsEmpty, permsNoAdmin, permsXRWAD},
		func(nameSuffix string, perms access.Permissions) error {
			d := s.DatabaseForId(wire.Id{Blessing: "root", Name: "db_" + nameSuffix}, nil)
			return d.Create(ctx, perms)
		})

	testPermsValidationOp(t, "c.Create()",
		[]access.Permissions{nil /* infer */, permsAdminOnly, permsComplex, permsRA, permsRWA},
		[]access.Permissions{permsEmpty, permsNoAdmin, permsXRWA, permsXRWAD},
		func(nameSuffix string, perms access.Permissions) error {
			c := rd.CollectionForId(wire.Id{Blessing: "root", Name: "cx_" + nameSuffix})
			return c.Create(ctx, perms)
		})

	testPermsValidationOp(t, "sg.Create()",
		[]access.Permissions{permsAdminOnly, permsComplex, permsRA},
		[]access.Permissions{nil, permsEmpty, permsNoAdmin, permsRWA, permsXRWA, permsXRWAD},
		func(nameSuffix string, perms access.Permissions) error {
			sg := rd.SyncgroupForId(wire.Id{Blessing: "root", Name: "sg_" + nameSuffix})
			return sg.Create(ctx, wire.SyncgroupSpec{
				Collections: []wire.Id{rc.Id()},
				Perms:       perms,
			}, wire.SyncgroupMemberInfo{SyncPriority: 42})
		})

	testPermsValidationOp(t, "s.SetPermissions()",
		[]access.Permissions{permsAdminOnly, permsComplex, permsRA, permsRWA, permsXRWA, permsXRWAD},
		[]access.Permissions{nil, permsEmpty, permsNoAdmin},
		func(_ string, perms access.Permissions) error {
			return s.SetPermissions(ctx, perms, "")
		})

	testPermsValidationOp(t, "d.SetPermissions()",
		[]access.Permissions{permsAdminOnly, permsComplex, permsRA, permsRWA, permsXRWA},
		[]access.Permissions{nil, permsEmpty, permsNoAdmin, permsXRWAD},
		func(_ string, perms access.Permissions) error {
			return rd.SetPermissions(ctx, perms, "")
		})

	testPermsValidationOp(t, "c.SetPermissions()",
		[]access.Permissions{permsAdminOnly, permsComplex, permsRA, permsRWA},
		[]access.Permissions{nil, permsEmpty, permsNoAdmin, permsXRWA, permsXRWAD},
		func(_ string, perms access.Permissions) error {
			return rc.SetPermissions(ctx, perms)
		})

	testPermsValidationOp(t, "sg.SetSpec()",
		[]access.Permissions{permsAdminOnly, permsComplex, permsRA},
		[]access.Permissions{nil, permsEmpty, permsNoAdmin, permsRWA, permsXRWA, permsXRWAD},
		func(_ string, perms access.Permissions) error {
			return rsg.SetSpec(ctx, wire.SyncgroupSpec{
				Collections: []wire.Id{rc.Id()},
				Perms:       perms,
			}, "")
		})
}

func testPermsValidationOp(t *testing.T, desc string, validPerms, invalidPerms []access.Permissions, op func(nameSuffix string, perms access.Permissions) error) {
	for i, p := range validPerms {
		nameSuffix := fmt.Sprintf("ok_%02d", i)
		if err := op(nameSuffix, p); err != nil {
			t.Errorf("perms %v: %s failed: %v", p, desc, err)
		}
	}

	for i, p := range invalidPerms {
		nameSuffix := fmt.Sprintf("notok_%02d", i)
		if err := op(nameSuffix, p); err == nil {
			t.Errorf("perms %v: %s should have failed", p, desc)
		}
	}
}
