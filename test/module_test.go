package test

import (
	"fmt"
	"testing"
	"weqi_service/api"
	"weqi_service/models"
	"weqi_service/wxpay"
)

func TestModule(t *testing.T) {
	module := models.ModuleInstance{
		InstanceID: 1,
		ModuleID: 1,
		ModuleName: "jxc",
		UserID: 1,
		Admin: "alex",
		ComID: 1,
		Company: "facebook",
		Domains: []string{"www.example1.com", "www.example2.com"},
	}
	api.AddRecordToRemoteModule("http://127.0.0.1:3000/add_superadmin", module)
}

func TestRand(t *testing.T) {
	code := api.GenRandomDigitCode(6)
	fmt.Println(code)
}

func TestSignKey(t *testing.T) {
	wxpay.GetTestSignKey()
}