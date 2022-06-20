package rt

import "log"

func CheckNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
