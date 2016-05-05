// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was auto-generated via go generate.
// DO NOT UPDATE MANUALLY

/*
Command vdltestgen generates types and values for the vdltest package.  The
following files are generated:

   vtype_gen.vdl       - A variety of types useful for testing.
   ventry_pass_gen.vdl - Entries that pass conversion from source to target.
   ventry_fail_gen.vdl - Entries that fail conversion from source to target.

This tool does not run the vdl tool on the generated *.vdl files; you must do
that yourself, typically via "jiri go install".

Instead of running this tool manually, it is typically invoked via:

   $ jiri run go generate v.io/v23/vdl/vdltest

Usage:
   vdltestgen [flags]

The vdltestgen flags are:
 -ventry-fail=ventry_fail_gen.vdl
   Name of the generated ventry fail file, containing VDL values that fail
   conversion tests.
 -ventry-pass=ventry_pass_gen.vdl
   Name of the generated ventry pass file, containing VDL values that pass
   conversion tests.
 -vtype=vtype_gen.vdl
   Name of the generated vtype file, containing VDL types.

The global flags are:
 -metadata=<just specify -metadata to activate>
   Displays metadata for the program and exits.
 -time=false
   Dump timing information to stderr before exiting the program.
 -vdltest=
   Filter vdltest.All to only return entries that contain the given substring.
*/
package main