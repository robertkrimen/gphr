package gphr

import (
	"testing"

	"./terst"
)

var is = terst.Is

func TestBinary(t *testing.T) {
	terst.Terst(t, func() {
		bn := NewBinary("../../example/example_linux_386")
		is(bn.Path, "../../example/example_linux_386")
		is(bn.Name, "example_linux_386")
		is(bn.Program, "example")
		is(bn.GOOS, "linux")
		is(bn.GOARCH, "386")

		is(bn.Identifier(), "example-linux-386")
		is(bn.Dash(), "example-linux-386")
		is(bn.Underscore(), "example_linux_386")

		is(bn.Match("example_linux_386"), true)
		is(bn.Match("example_linux_amd64"), false)

		bn = NewBinary("example_linux_386")
		is(bn.Path, "example_linux_386")
		is(bn.Name, "example_linux_386")
		is(bn.Program, "example")
		is(bn.GOOS, "linux")
		is(bn.GOARCH, "386")

		is(bn.Match("example_linux_386"), true)
	})
}
