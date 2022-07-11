package main

import (
	"fmt"
	"github.com/morganleroi/AzBlobStorage/deploy"
)

func main() {
	deployPackage, err := deploy.GetDeployedPackage("yamaalgolia",
		"XXX")
	if err != nil {
		fmt.Printf("Ouppsss %s", err)
		return
	}

	fmt.Println(deployPackage)
}
