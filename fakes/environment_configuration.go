package fakes

import (
	"sync"

	packit "github.com/paketo-buildpacks/packit/v2"
)

type EnvironmentConfiguration struct {
	ConfigureCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Layer         packit.Layer
			ExtensionsDir string
			DefaultIni    string
			ScanDirs      []string
		}
		Returns struct {
			Error error
		}
		Stub func(packit.Layer, string, string, []string) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 packit.Layer, param2 string, param3 string, param4 []string) error {
	f.ConfigureCall.mutex.Lock()
	defer f.ConfigureCall.mutex.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.Layer = param1
	f.ConfigureCall.Receives.ExtensionsDir = param2
	f.ConfigureCall.Receives.DefaultIni = param3
	f.ConfigureCall.Receives.ScanDirs = param4
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1, param2, param3, param4)
	}
	return f.ConfigureCall.Returns.Error
}
