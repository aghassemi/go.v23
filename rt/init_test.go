package rt_test

import (
	"net"
	"testing"

	"veyron2"
	"veyron2/config"
	"veyron2/rt"
	"veyron2/security"
)

func ExampleInit() {
	r := rt.Init()
	// Go ahead and use the runtime.
	log := r.Logger()
	log.Infof("hello world")
}

type myprofile struct{}

func (mp *myprofile) Name() string {
	return "test"
}

func (mp *myprofile) Runtime() string {
	return ""
}

func (mp *myprofile) AddressChooser() veyron2.AddressChooser {
	return func(network string, addrs []net.Addr) (net.Addr, error) {
		return &net.IPAddr{IP: net.ParseIP("127.0.0.1")}, nil
	}
}

func (mp *myprofile) Platform() *veyron2.Platform {
	id := security.FakePublicID("anyoldid")
	return &veyron2.Platform{"google", id, "v1", "any", "rel1", ".2", "who knows", "this host"}
}

func (mp *myprofile) String() string {
	return "myprofile on " + mp.Platform().String()
}

func (mp *myprofile) Init(veyron2.Runtime, *config.Publisher) {}

func ExampleInitWithProfile() {
	r := rt.Init(veyron2.ProfileOpt{&myprofile{}})
	// Go ahead and use the runtime.
	log := r.Logger()
	log.Infof("hello world from my product: %s", r.Profile())
}

// TODO(cnicolaou): add tests to:
//  - catch mismatched profile and runtimes - e.g. profile asks for "foo"
// runtime, but only bar is available.
//  - tests to catch multiple calls to init with different options

func TestErrorOnNew(t *testing.T) {
	_, err := rt.New(veyron2.RuntimeOpt{"foo"})
	if err == nil {
		t.Errorf("expected an error!")
	}
}
