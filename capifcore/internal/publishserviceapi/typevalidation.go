package publishserviceapi

import (
	"errors"
	//"fmt"
	"strings"
)

func (sd ServiceAPIDescription) Validate() error {
	if len(strings.TrimSpace(sd.ApiName)) == 0 {
		return errors.New("ServiceAPIDescription missing required apiName")
	}
	return nil
}
