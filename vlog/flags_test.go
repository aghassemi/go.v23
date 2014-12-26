package vlog_test

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"v.io/core/veyron/lib/modules"
	"v.io/core/veyron/lib/testutil"

	"v.io/core/veyron2/vlog"
)

func TestHelperProcess(t *testing.T) {
	modules.DispatchInTest()
}

func init() {
	testutil.Init()
	modules.RegisterChild("child", "", child)
}

func child(stdin io.Reader, stdout, stderr io.Writer, env map[string]string, args ...string) error {
	tmp := filepath.Join(os.TempDir(), "foo")
	flag.Set("log_dir", tmp)
	flag.Set("vmodule", "foo=2")
	flags := vlog.Log.ExplicitlySetFlags()
	if v, ok := flags["log_dir"]; !ok || v != tmp {
		return fmt.Errorf("log_dir was supposed to be %v", tmp)
	}
	if v, ok := flags["vmodule"]; !ok || v != "foo=2" {
		return fmt.Errorf("vmodule was supposed to be foo=2")
	}
	if f := flag.Lookup("max_stack_buf_size"); f == nil {
		return fmt.Errorf("max_stack_buf_size is not a flag")
	}
	maxStackBufSizeSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "max_stack_buf_size" {
			maxStackBufSizeSet = true
		}
	})
	if v, ok := flags["max_stack_buf_size"]; ok && !maxStackBufSizeSet {
		return fmt.Errorf("max_stack_buf_size unexpectedly set to %v", v)
	}
	return nil
}

func TestFlags(t *testing.T) {
	sh, err := modules.NewShell(nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer sh.Cleanup(nil, nil)
	h, err := sh.Start("child", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err = h.Shutdown(nil, os.Stderr); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
