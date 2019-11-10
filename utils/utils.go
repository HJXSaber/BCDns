package utils

import "os"

func DBExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// merge map1 into map2
func CoverMap(map1, map2 map[string]string) map[string]string {
	for k, v := range map1 {
		map2[k] = v
	}
	return map2
}
