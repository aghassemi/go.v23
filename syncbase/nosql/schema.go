// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nosql

import (
	wire "v.io/syncbase/v23/services/syncbase/nosql"
	"v.io/v23/context"
)

// NewSchema creates a new Schema object.
func NewSchema(metadata wire.SchemaMetadata, upgrader SchemaUpgrader) *Schema {
	return &Schema{
		Metadata: metadata,
		Upgrader: upgrader,
	}
}

//////////////////////////////////////////
// Implementation of SchemaManager (Not part of public client API)

type schemaManagerImpl struct {
	c wire.DatabaseClientMethods
}

func newSchemaManager(client wire.DatabaseClientMethods) schemaManagerImpl {
	return schemaManagerImpl{
		c: client,
	}
}

// GetSchemaMetadata retrieves the schema metadata for the database it is
// derived from.
func (sm *schemaManagerImpl) getSchemaMetadata(ctx *context.T) (wire.SchemaMetadata, error) {
	return sm.c.GetSchemaMetadata(ctx)
}

// SetSchemaMetadata stores the schema metadata for the database it is
// derived from.
func (sm *schemaManagerImpl) setSchemaMetadata(ctx *context.T, metadata wire.SchemaMetadata) error {
	return sm.c.SetSchemaMetadata(ctx, metadata)
}
