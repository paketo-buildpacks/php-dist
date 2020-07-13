package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/packit"
)

type EnvironmentConfiguration struct {
	ConfigureCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Layer packit.Layer
		}
		Returns struct {
			Error error
		}
		Stub func(packit.Layer) error
	}
}

func (f *EnvironmentConfiguration) Configure(param1 packit.Layer) error {
	f.ConfigureCall.Lock()
	defer f.ConfigureCall.Unlock()
	f.ConfigureCall.CallCount++
	f.ConfigureCall.Receives.Layer = param1
	if f.ConfigureCall.Stub != nil {
		return f.ConfigureCall.Stub(param1)
	}
	return f.ConfigureCall.Returns.Error
}
