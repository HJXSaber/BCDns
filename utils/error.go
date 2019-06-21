package utils

import "log"

func Assert(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}