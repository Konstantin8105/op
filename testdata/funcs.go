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
	const op = "st.Func2"
	return
}

func (s *st) func3() (err error) {
	const op = "st.func3"
	return
}

func (s *st) Func4(f2 string) (err error) {
	const op = "st.Func4"
	return
}

func main() {

}
