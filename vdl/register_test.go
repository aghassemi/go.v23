package vdl

import (
	"reflect"
	"sync"
	"testing"
)

var NUnionWant = []*reflectInfo{{
	Type: reflect.TypeOf((*NUnionABC)(nil)).Elem(),
	Name: "v.io/core/veyron2/vdl.NUnionABC",
	UnionFields: []reflectField{
		{"A", reflect.TypeOf(false), reflect.TypeOf(NUnionABCA{})},
		{"B", reflect.TypeOf(string("")), reflect.TypeOf(NUnionABCB{})},
		{"C", reflect.TypeOf(NStructInt64{}), reflect.TypeOf(NUnionABCC{})},
	},
}}

var reflectInfoTests = []struct {
	rt reflect.Type
	ri []*reflectInfo
}{
	{reflect.TypeOf(int64(0)), []*reflectInfo{{Type: reflect.TypeOf(int64(0))}}},
	{reflect.TypeOf(string("")), []*reflectInfo{{Type: reflect.TypeOf(string(""))}}},
	{reflect.TypeOf([]byte{}), []*reflectInfo{{Type: reflect.TypeOf([]byte{})}}},
	{
		reflect.TypeOf(NEnumA),
		[]*reflectInfo{{
			Type:       reflect.TypeOf(NEnumA),
			Name:       "v.io/core/veyron2/vdl.NEnum",
			EnumLabels: []string{"A", "B", "C", "ABC"},
		}},
	},
	{reflect.TypeOf((*NUnionABC)(nil)).Elem(), NUnionWant},
	{reflect.TypeOf(NUnionABCA{}), NUnionWant},
	{reflect.TypeOf(NUnionABCB{}), NUnionWant},
	{reflect.TypeOf(NUnionABCC{}), NUnionWant},
	{
		reflect.TypeOf(NRecurseSelf{}),
		[]*reflectInfo{{
			Type: reflect.TypeOf(NRecurseSelf{}),
			Name: "v.io/core/veyron2/vdl.NRecurseSelf",
		}},
	},
	{
		reflect.TypeOf(NRecurseA{}),
		[]*reflectInfo{
			{
				Type: reflect.TypeOf(NRecurseA{}),
				Name: "v.io/core/veyron2/vdl.NRecurseA",
			},
			{
				Type: reflect.TypeOf(NRecurseB{}),
				Name: "v.io/core/veyron2/vdl.NRecurseB",
			},
		},
	},
	{
		reflect.TypeOf(NRecurseB{}),
		[]*reflectInfo{
			{
				Type: reflect.TypeOf(NRecurseB{}),
				Name: "v.io/core/veyron2/vdl.NRecurseB",
			},
			{
				Type: reflect.TypeOf(NRecurseA{}),
				Name: "v.io/core/veyron2/vdl.NRecurseA",
			},
		},
	},
}

// Test deriveReflectInfo success.
func TestDeriveReflectInfo(t *testing.T) {
	for _, test := range reflectInfoTests {
		ri, err := deriveReflectInfo(test.rt)
		if ri == nil || err != nil {
			t.Errorf("%s deriveReflectInfo failed: (%v, %v)", test.rt, ri, err)
			continue
		}
		if got, want := ri, test.ri[0]; !reflect.DeepEqual(got, want) {
			t.Errorf("%s got %v, want %v", test.rt, got, want)
		}
	}
}

// Test Register called by multiple goroutines concurrently on the same types,
// to expose locking issues in the registry.
func TestRegister(t *testing.T) {
	var done sync.WaitGroup
	for i := 0; i < 3; i++ {
		done.Add(1)
		go func() {
			testRegister(t)
			done.Done()
		}()
	}
	done.Wait()
}

func testRegister(t *testing.T) {
	for _, test := range reflectInfoTests {
		Register(reflect.New(test.rt).Interface())
		for _, testri := range test.ri {
			if testri.Name != "" {
				ri := reflectInfoFromName(testri.Name)
				if got, want := ri, testri; !reflect.DeepEqual(got, want) {
					t.Errorf("%s reflectInfoFromName got %v, want %v", test.rt, got, want)
				}
			}
		}
	}
}

type (
	nBadDescribe1 struct{}
	nBadDescribe2 struct{}
	nBadDescribe3 struct{}

	nBadEnumNoLabels int
	nBadEnumString1  int
	nBadEnumString2  int
	nBadEnumString3  int
	nBadEnumSet1     int
	nBadEnumSet2     int
	nBadEnumSet3     int
	nBadEnumSet4     int

	nBadUnionNoFields struct{}
	nBadUnionUnexp    struct{}
	nBadUnionField1   struct{}
	nBadUnionField2   struct{}
	nBadUnionField3   struct{}
	nBadUnionName1    struct{ Value bool }
	nBadUnionName2    struct{ Value bool }
)

// No description
func (nBadDescribe1) __VDLReflect() { panic("X") }

// In-arg isn't a struct
func (nBadDescribe2) __VDLReflect(int) { panic("X") }

// Can't have out-arg
func (nBadDescribe3) __VDLReflect(struct{}) error { panic("X") }

// No enum labels
func (nBadEnumNoLabels) __VDLReflect(struct{ Enum struct{} }) { panic("X") }

// No String method
func (nBadEnumString1) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }

// String method isn't String() string
func (nBadEnumString2) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumString2) String()                                        { panic("X") }

// String method isn't String() string
func (nBadEnumString3) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumString3) String() bool                                   { panic("X") }

// No Set method
func (nBadEnumSet1) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumSet1) String() string                                 { panic("X") }

// Set method isn't Set(string) error
func (nBadEnumSet2) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumSet2) String() string                                 { panic("X") }
func (nBadEnumSet2) Set()                                           { panic("X") }

// Set method isn't Set(string) error
func (nBadEnumSet3) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumSet3) String() string                                 { panic("X") }
func (nBadEnumSet3) Set(bool) error                                 { panic("X") }

// Set method receiver isn't a pointer
func (nBadEnumSet4) __VDLReflect(struct{ Enum struct{ A string } }) { panic("X") }
func (nBadEnumSet4) String() string                                 { panic("X") }
func (nBadEnumSet4) Set(string) error                               { panic("X") }

// No union fields
func (nBadUnionNoFields) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{}
}) {
	panic("X")
}

// Field name isn't exported
func (nBadUnionUnexp) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ a NUnionABCA }
}) {
	panic("X")
}

// Field type isn't struct
func (nBadUnionField1) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ A bool }
}) {
	panic("X")
}

// Field type has no field
func (nBadUnionField2) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ A struct{} }
}) {
	panic("X")
}

// Field type name isn't "Value"
func (nBadUnionField3) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ A struct{ value bool } }
}) {
	panic("X")
}

// Name method isn't Name() string
func (nBadUnionName1) Name() { panic("X") }
func (nBadUnionName1) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ A nBadUnionName1 }
}) {
	panic("X")
}

// Name method isn't Name() string
func (nBadUnionName2) Name() bool { panic("X") }
func (nBadUnionName2) __VDLReflect(struct {
	Type  NUnionABC
	Union struct{ A nBadUnionName2 }
}) {
	panic("X")
}

// rtErrorTest describes a test case with rt as input, and errstr as output.
type rtErrorTest struct {
	rt     reflect.Type
	errstr string
}

const (
	badDescribe   = `invalid __VDLReflect (want __VDLReflect(struct{...}))`
	badEnumString = `must have method String() string`
	badEnumSet    = `must have pointer method Set(string) error`
	badUnionField = `bad concrete field type`
	badUnionName  = `must have method Name() string`
)

var reflectInfoErrorTests = []rtErrorTest{
	{reflect.TypeOf(nBadDescribe1{}), badDescribe},
	{reflect.TypeOf(nBadDescribe2{}), badDescribe},
	{reflect.TypeOf(nBadDescribe3{}), badDescribe},
	{reflect.TypeOf(nBadEnumNoLabels(0)), `no labels`},
	{reflect.TypeOf(nBadEnumString1(0)), badEnumString},
	{reflect.TypeOf(nBadEnumString2(0)), badEnumString},
	{reflect.TypeOf(nBadEnumString3(0)), badEnumString},
	{reflect.TypeOf(nBadEnumSet1(0)), badEnumSet},
	{reflect.TypeOf(nBadEnumSet2(0)), badEnumSet},
	{reflect.TypeOf(nBadEnumSet3(0)), badEnumSet},
	{reflect.TypeOf(nBadEnumSet4(0)), badEnumSet},
	{reflect.TypeOf(nBadUnionNoFields{}), `no fields`},
	{reflect.TypeOf(nBadUnionUnexp{}), `must be exported`},
	{reflect.TypeOf(nBadUnionField1{}), badUnionField},
	{reflect.TypeOf(nBadUnionField2{}), badUnionField},
	{reflect.TypeOf(nBadUnionField3{}), badUnionField},
	{reflect.TypeOf(nBadUnionName1{}), badUnionName},
	{reflect.TypeOf(nBadUnionName2{}), badUnionName},
}

// Test deriveReflectInfo errors.
func TestDeriveReflectInfoError(t *testing.T) {
	for _, test := range reflectInfoErrorTests {
		got, err := deriveReflectInfo(test.rt)
		ExpectErr(t, err, test.errstr, "deriveReflectInfo(%v)", test.rt)
		if got != nil {
			t.Errorf("deriveReflectInfo(%v) got %v, want nil", test.rt, got)
		}
	}
}
