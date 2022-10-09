package kitx

import "fmt"

const (
	version = "v0.0.2"
	logo    = `
							  _      _               
							 | |    (_)   _          
	  ____   ___     _____   | |  _  _  _| |_  _   _ 
	 / _  | / _ \   (_____)  | |_/ )| |(_   _)( \ / )
	( (_| || |_| |           |  _ ( | |  | |_  ) X ( 
	 \___ | \___/            |_| \_)|_|   \__)(_/ \_)
	(_____|                                           
`
)

func startingPrint(id, name string) {
	fmt.Printf("%s \n", logo)
	fmt.Printf("\x1b[%dmKitx Version: %s\x1b[0m \n", 36, version)
	fmt.Printf("\x1b[%dmApp ID: %s\x1b[0m \n", 36, id)
	fmt.Printf("\x1b[%dmApp Name: %s\x1b[0m \n", 36, name)
	fmt.Printf("\x1b[%dmStarting App ...\x1b[0m \n", 34)
	fmt.Println("")
}
