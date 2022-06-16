// generate init data sql, for manual init db
package main

import (
	"fmt"

	"github.com/hwiewie/APIServer/database/initial"
)

func main() {
	for _, data := range initial.InitialData {
		fmt.Println(data)
	}
}