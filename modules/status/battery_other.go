// +build !openbsd,!freebsd,!netbsd

package status

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func (b *battery) value() (float64, error) {
	nowFile, fullFile := pickChargeOrEnergy()
	fullStr, err1 := ioutil.ReadFile(fullFile)
	if os.IsNotExist(err1) {
		return 0, err1 // return quietly if the file was not present (desktop?)
	}
	nowStr, err2 := ioutil.ReadFile(nowFile)
	if err1 != nil || err2 != nil {
		log.Println("Error reading battery info", err1)
		return 0, err1
	}

	now, err1 := strconv.Atoi(strings.TrimSpace(string(nowStr)))
	full, err2 := strconv.Atoi(strings.TrimSpace(string(fullStr)))
	if err1 != nil || err2 != nil {
		log.Println("Error converting battery info", err1)
		return 0, err1
	}

	return float64(now) / float64(full), nil
}
