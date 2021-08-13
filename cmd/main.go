package main

import (
	"fmt"
	"runtime"

	"github.com/chindeo/pkg/file"
)

func main() {
	abpath := file.GetCurrentAbPath()
	abexecpath := file.GetCurrentAbPathByExecutable()
	abcalpath := file.GetCurrentAbPathByCaller()
	// funcName := file.GetCurrentFuncNameByCaller()
	fmt.Println(abpath)
	fmt.Println(abexecpath)
	fmt.Println(abcalpath)
	fmt.Println(runtime.Caller(0))
	file.WriteString("C:/Users/Administrator/go/src/github.com/chindeo/pkg/cmd/path.json", fmt.Sprintf("%s\n%s\n%s\n", abpath, abexecpath, abcalpath))
}
