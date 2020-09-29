package atexit

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	require := require.New(t)
	current := len(handlers)
	Register(func() {})
	require.Equal(current+1, len(handlers), "register")
}

func TestCancel(t *testing.T) {
	require := require.New(t)

	id := Register(func() {})
	require.NoError(id.Cancel())
	_, ok := handlers[id]
	require.False(ok, "cancel")
}

func TestHandler(t *testing.T) {
	require := require.New(t)
	root, err := ioutil.TempDir("", "atexit-test")
	require.NoError(err, "TempDir")

	progFile := path.Join(root, "main.go")
	err = ioutil.WriteFile(progFile, []byte(testprog), 0666)
	require.NoError(err, "prog")

	here, err := filepath.Abs(".")
	require.NoError(err, "abs .")

	mod := fmt.Sprintf(modTmpl, here)
	modFile := path.Join(root, "go.mod")
	err = ioutil.WriteFile(modFile, []byte(mod), 0666)
	require.NoError(err, "mod")
	outFile := path.Join(root, "main.out")

	arg := time.Now().UTC().String()
	err = exec.Command("go", "run", progFile, outFile, arg).Run()
	require.Error(err, "run")

	data, err := ioutil.ReadFile(outFile)
	require.NoError(err, "read out")
	require.Equal(arg, string(data), "output")
}

var (
	testprog = `
// Test program for atexit, gets output file and data as arguments and writes
// data to output file in atexit handler.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/tebeka/atexit"
)

var outfile = ""
var data = ""

func handler() {
	ioutil.WriteFile(outfile, []byte(data), 0666)
}

func badHandler() {
	n := 0
	fmt.Println(1/n)
}

func unusedHandler() {
	ioutil.WriteFile(outfile, []byte("\nunused"), 0666)
}

func main() {
	flag.Parse()
	outfile = flag.Arg(0)
	data = flag.Arg(1)

	atexit.Register(handler)
	id := atexit.Register(unusedHandler)
	atexit.Register(badHandler)
	id.Cancel()
	atexit.Exit(1)
}
`
	modTmpl = `
module testexit

go 1.13

replace github.com/tebeka/atexit => %s
`
)
