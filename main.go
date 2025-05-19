package main

import (
	"fmt"
	"os"

	"github.com/littleworks-inc/cloudcost/cmd"
)

func main() {
	fmt.Println(`
 ______     __         ______     __  __     _____     ______     ______     ______     ______  
/\  ___\   /\ \       /\  __ \   /\ \/\ \   /\  __-.  /\  ___\   /\  __ \   /\  ___\   /\__  _\ 
\ \ \____  \ \ \____  \ \ \/\ \  \ \ \_\ \  \ \ \/\ \ \ \ \____  \ \ \/\ \  \ \___  \  \/_/\ \/ 
 \ \_____\  \ \_____\  \ \_____\  \ \_____\  \ \____-  \ \_____\  \ \_____\  \/\_____\    \ \_\ 
  \/_____/   \/_____/   \/_____/   \/_____/   \/____/   \/_____/   \/_____/   \/_____/     \/_/ 
                                                                                                 
Cloud Cost Estimator for Infrastructure-as-Code (v` + cmd.Version + `)
`)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
