// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syncbase

import (
	wire "v.io/syncbase/v23/services/syncbase"
	"v.io/syncbase/v23/syncbase/nosql"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/security/access"
)

type app struct {
	c            wire.AppClientMethods
	name         string
	relativeName string
}

var _ App = (*app)(nil)

// TODO(sadovsky): Validate names before sending RPCs.

// Name implements App.Name.
func (a *app) Name() string {
	return a.relativeName
}

// NoSQLDatabase implements App.NoSQLDatabase.
func (a *app) NoSQLDatabase(relativeName string) nosql.Database {
	name := naming.Join(a.name, relativeName)
	// TODO(sadovsky): It's annoying that we must export nosql.DatabaseImpl.
	return nosql.NewDatabase(name, relativeName)
}

// ListDatabases implements App.ListDatabases.
func (a *app) ListDatabases(ctx *context.T) ([]string, error) {
	// TODO(sadovsky): Implement on top of Glob.
	return nil, nil
}

// Create implements App.Create.
func (a *app) Create(ctx *context.T, perms access.Permissions) error {
	return a.c.Create(ctx, perms)
}

// Delete implements App.Delete.
func (a *app) Delete(ctx *context.T) error {
	return a.c.Delete(ctx)
}

// SetPermissions implements App.SetPermissions.
func (a *app) SetPermissions(ctx *context.T, perms access.Permissions, version string) error {
	return a.c.SetPermissions(ctx, perms, version)
}

// GetPermissions implements App.GetPermissions.
func (a *app) GetPermissions(ctx *context.T) (perms access.Permissions, version string, err error) {
	return a.c.GetPermissions(ctx)
}
