package publishserviceapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	serviceDescriptionUnderTest := ServiceAPIDescription{}
	err := serviceDescriptionUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "apiName")
	}

	serviceDescriptionUnderTest.ApiName = "apiName"
	err = serviceDescriptionUnderTest.Validate()
	assert.Nil(t, err)

}
