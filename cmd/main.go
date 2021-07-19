package main

import (
	"fmt"

	"github.com/chindeo/pkg/file"
)

func main() {
	abpath := file.GetCurrentAbPath()
	abexecpath := file.GetCurrentAbPathByExecutable()
	abcalpath := file.GetCurrentAbPathByCaller()
	fmt.Println(abpath)
	fmt.Println(abexecpath)
	fmt.Println(abcalpath)
	file.WriteString("C:/Users/Administrator/go/src/github.com/chindeo/pkg/cmd/path.json", fmt.Sprintf("%s\n%s\n%s\n", abpath, abexecpath, abcalpath))
}
