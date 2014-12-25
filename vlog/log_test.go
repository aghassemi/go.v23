package vlog_test

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"v.io/veyron/veyron2/vlog"
)

func ExampleConfigure() {
	vlog.ConfigureLogger()
}

func ExampleInfo() {
	vlog.Info("hello")
}

func ExampleError() {
	vlog.Errorf("%s", "error")
	if vlog.V(2) {
		vlog.Info("some spammy message")
	}
	vlog.VI(2).Infof("another spammy message")
}

func readLogFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var contents []string
	for _, fi := range files {
		// Skip symlinks to avoid double-counting log lines.
		if !fi.Mode().IsRegular() {
			continue
		}
		file, err := os.Open(filepath.Join(dir, fi.Name()))
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if line := scanner.Text(); len(line) > 0 && line[0] == 'I' {
				contents = append(contents, line)
			}
		}
	}
	return contents, nil
}

func TestHeaders(t *testing.T) {
	dir, err := ioutil.TempDir("", "logtest")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	logger, err := vlog.NewLogger("testHeader", vlog.LogDir(dir), vlog.Level(2))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	logger.Infof("abc\n")
	logger.Infof("wombats\n")
	logger.VI(1).Infof("wombats again\n")
	logger.FlushLog()
	contents, err := readLogFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	fileRE := regexp.MustCompile(`\S+ \S+ \S+ (.*):.*`)
	for _, line := range contents {
		name := fileRE.FindStringSubmatch(line)
		if len(name) < 2 {
			t.Errorf("failed to find file in %s", line)
			continue
		}
		if name[1] != "log_test.go" {
			t.Errorf("unexpected file name: %s", name[1])
			continue
		}
	}
	if want, got := 3, len(contents); want != got {
		t.Errorf("Expected %d info lines, got %d instead", want, got)
	}
}

func myLoggedFunc() {
	f := vlog.LogCall("entry")
	f("exit")
}

func TestLogCall(t *testing.T) {
	dir, err := ioutil.TempDir("", "logtest")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	logger, err := vlog.NewLogger("testHeader", vlog.LogDir(dir), vlog.Level(2))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	saveLog := vlog.Log
	defer vlog.SetLog(saveLog)
	vlog.SetLog(logger)

	myLoggedFunc()
	vlog.FlushLog()
	contents, err := readLogFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	logCallLineRE := regexp.MustCompile(`\S+ \S+ \S+ ([^:]*):.*(call|return)\[(\S*)`)
	for _, line := range contents {
		match := logCallLineRE.FindStringSubmatch(line)
		if len(match) != 4 {
			t.Errorf("failed to match %s", line)
			continue
		}
		fileName, callType, funcName := match[1], match[2], match[3]
		if fileName != "log_test.go" {
			t.Errorf("unexpected file name: %s", fileName)
			continue
		}
		if callType != "call" && callType != "return" {
			t.Errorf("unexpected call type: %s", callType)
		}
		if funcName != "vlog_test.myLoggedFunc" {
			t.Errorf("unexpected func name: %s", funcName)
		}
	}
	if want, got := 2, len(contents); want != got {
		t.Errorf("Expected %d info lines, got %d instead", want, got)
	}
}

func TestVModule(t *testing.T) {
	dir, err := ioutil.TempDir("", "logtest")
	defer os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	logger, err := vlog.NewLogger("testVmodule", vlog.LogDir(dir))
	if logger.V(2) || logger.V(3) {
		t.Errorf("Logging should not be enabled at levels 2 & 3")
	}
	spec := vlog.ModuleSpec{}
	if err := spec.Set("*log_test=2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := logger.ConfigureLogger(spec); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !logger.V(2) {
		t.Errorf("logger.V(2) should be true")
	}
	if logger.V(3) {
		t.Errorf("logger.V(3) should be false")
	}
	if vlog.V(2) || vlog.V(3) {
		t.Errorf("Logging should not be enabled at levels 2 & 3")
	}
	vlog.Log.ConfigureLogger(spec)
	if !vlog.V(2) {
		t.Errorf("vlog.V(2) should be true")
	}
	if vlog.V(3) {
		t.Errorf("vlog.V(3) should be false")
	}
	if vlog.VI(2) != vlog.Log {
		t.Errorf("vlog.V(2) should be vlog.Log")
	}
	if vlog.VI(3) == vlog.Log {
		t.Errorf("vlog.V(3) should not be vlog.Log")
	}
}
