// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(sadovsky): Enable once the server implementation has been updated to
// reflect the new Syncbase design.
// +build ignore

package syncbase_test

// Note: v.io/x/ref/services/groups/internal/server/server_test.go has some
// helpful code snippets to model after.

import (
	"reflect"
	"runtime/debug"
	"testing"

	"v.io/syncbase/v23/syncbase"
	"v.io/syncbase/v23/syncbase/common"
	"v.io/syncbase/x/ref/services/syncbase/server"
	"v.io/syncbase/x/ref/services/syncbase/store/memstore"
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/security"
	"v.io/v23/security/access"
	"v.io/x/lib/vlog"
	_ "v.io/x/ref/profiles"
	tsecurity "v.io/x/ref/test/testutil"
)

func defaultPermissions() access.Permissions {
	perms := access.Permissions{}
	for _, tag := range access.AllTypicalTags() {
		perms.Add(security.BlessingPattern("server/client"), string(tag))
	}
	return perms
}

func newServer(ctx *context.T, perms access.Permissions) (string, func()) {
	s, err := v23.NewServer(ctx)
	if err != nil {
		vlog.Fatal("v23.NewServer() failed: ", err)
	}
	eps, err := s.Listen(v23.GetListenSpec(ctx))
	if err != nil {
		vlog.Fatal("s.Listen() failed: ", err)
	}

	service := server.NewService(memstore.New())
	if perms == nil {
		perms = defaultPermissions()
	}
	if err := service.Create(perms); err != nil {
		vlog.Fatal("service.Create() failed: ", err)
	}
	d := server.NewDispatcher(service)

	if err := s.ServeDispatcher("", d); err != nil {
		vlog.Fatal("s.ServeDispatcher() failed: ", err)
	}

	name := naming.JoinAddressName(eps[0].String(), "")
	return name, func() {
		s.Stop()
	}
}

func setupOrDie(perms access.Permissions) (clientCtx *context.T, serverName string, cleanup func()) {
	ctx, shutdown := v23.Init()
	cp, sp := tsecurity.NewPrincipal("client"), tsecurity.NewPrincipal("server")

	// Have the server principal bless the client principal as "client".
	blessings, err := sp.Bless(cp.PublicKey(), sp.BlessingStore().Default(), "client", security.UnconstrainedUse())
	if err != nil {
		vlog.Fatal("sp.Bless() failed: ", err)
	}
	// Have the client present its "client" blessing when talking to the server.
	if _, err := cp.BlessingStore().Set(blessings, "server"); err != nil {
		vlog.Fatal("cp.BlessingStore().Set() failed: ", err)
	}
	// Have the client treat the server's public key as an authority on all
	// blessings that match the pattern "server".
	if err := cp.AddToRoots(blessings); err != nil {
		vlog.Fatal("cp.AddToRoots() failed: ", err)
	}

	clientCtx, err = v23.SetPrincipal(ctx, cp)
	if err != nil {
		vlog.Fatal("v23.SetPrincipal() failed: ", err)
	}
	serverCtx, err := v23.SetPrincipal(ctx, sp)
	if err != nil {
		vlog.Fatal("v23.SetPrincipal() failed: ", err)
	}

	serverName, stopServer := newServer(serverCtx, perms)
	cleanup = func() {
		stopServer()
		shutdown()
	}
	return
}

func getPermissionsOrDie(ac common.AccessController, ctx *context.T, t *testing.T) access.Permissions {
	perms, _, err := ac.GetPermissions(ctx)
	if err != nil {
		debug.PrintStack()
		t.Fatalf("GetPermissions failed: %s", err)
	}
	return perms
}

////////////////////////////////////////
// Test cases

// TODO(sadovsky): Finish writing tests.

func TestNameAndKey(t *testing.T) {
	a := syncbase.NewService("s").App("a")

	if a.Name() != "a" {
		t.Errorf("Wrong name: %s", a.Name())
	}
}

// Tests that App.Create works as expected, including default Permissions.
func TestCreate(t *testing.T) {
	ctx, sName, cleanup := setupOrDie(nil)
	defer cleanup()

	s := syncbase.NewService(sName)
	a := s.App("a")
	d := a.NoSQLDatabase("d")

	// Database.Create should fail, since App does not exist.
	if err := d.Create(ctx, nil); err == nil {
		t.Fatalf("d.Create() should have failed")
	}

	// Create the App.
	// TODO(sadovsky): Test auth check (against Service Permissions).
	if err := a.Create(ctx, nil); err != nil {
		t.Fatalf("a.Create() failed: %s", err)
	}
	if wantPermissions, gotPermissions := defaultPermissions(), getPermissionsOrDie(a, ctx, t); !reflect.DeepEqual(wantPermissions, gotPermissions) {
		t.Errorf("Permissions do not match: want %v, got %v", wantPermissions, gotPermissions)
	}

	// Database.Create should now succeed.
	if err := d.Create(ctx, nil); err != nil {
		t.Fatalf("d.Create() failed: %s", err)
	}

	// Test App.Create with non-default Permissions.
	a2 := s.App("a2")
	perms := access.Permissions{}
	perms.Add(security.BlessingPattern("server/client"), string(access.Admin))
	if err := a2.Create(ctx, perms); err != nil {
		t.Fatalf("a2.Create() failed: %s", err)
	}
	if wantPermissions, gotPermissions := perms, getPermissionsOrDie(a2, ctx, t); !reflect.DeepEqual(wantPermissions, gotPermissions) {
		t.Errorf("Permissions do not match: want %v, got %v", wantPermissions, gotPermissions)
	}
}

// Tests that App.Delete works as expected.
func TestDelete(t *testing.T) {
}

// Tests that the various {Set,Get}Permissions methods work as expected.
func TestPermissionsMethods(t *testing.T) {
}
