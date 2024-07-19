package op_test

import (
	"fmt"
	"testing"

	"github.com/Konstantin8105/op"
)

type mockTest struct {
	log string
	res error
}

func (m *mockTest) Errorf(format string, args ...any) {
	m.res = fmt.Errorf(format, args...)
}

func (m *mockTest) Logf(format string, args ...any) {
	m.log += fmt.Sprintf(format, args...)
}

func (m mockTest) String() string {
	return fmt.Sprintf("Error: %v\nLog: %s", m.res, m.log)
}

func Test(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		op.Test(t, "op.go")
	})
	t.Run("not.valid", func(t *testing.T) {
		var m mockTest
		op.Test(&m, "testdata/funcs.go")
		if m.res == nil {
			t.Fatalf("not found error")
		}
		t.Logf("%s", m)
	})
}
