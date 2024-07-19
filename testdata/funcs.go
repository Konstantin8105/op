//go:build ignore

package main

func func1() (err error) {
	const op = "func1"
	return
}

func Func2(f2 string) (err error) {
	const op = "Func2"
	return
}

type st struct{}

func (s st) func1() (err error) {
	const op = "st.func1"
	return
}

func (s st) Func2(f2 string) (err error) {
	op = "st.Func2"
	return
}

func (s *st) func3() (err error) {
	const op = "st.func3"
	return
}

func (s *st) Func4(f2 string) (err error) {
	const op = "*st.Func4"
	return
}

func (s st) none() {}

func (s st) full() {
	const oper = "st.full"
	var f func()
	f = func() {}
	_ = f
}

func main() {
	var f func()
	f = func() {}
	_ = f
}
