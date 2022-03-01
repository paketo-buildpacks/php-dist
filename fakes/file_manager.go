package fakes

import (
	"sync"

	phpdist "github.com/paketo-buildpacks/php-dist"
)

type FileManager struct {
	FindExtensionsCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			LayerRoot string
		}
		Returns struct {
			String string
			Error  error
		}
		Stub func(string) (string, error)
	}
	WriteConfigCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			LayerRoot string
			CnbPath   string
			Data      phpdist.PhpIniConfig
		}
		Returns struct {
			DefaultConfig   string
			BuildpackConfig string
			Err             error
		}
		Stub func(string, string, phpdist.PhpIniConfig) (string, string, error)
	}
}

func (f *FileManager) FindExtensions(param1 string) (string, error) {
	f.FindExtensionsCall.mutex.Lock()
	defer f.FindExtensionsCall.mutex.Unlock()
	f.FindExtensionsCall.CallCount++
	f.FindExtensionsCall.Receives.LayerRoot = param1
	if f.FindExtensionsCall.Stub != nil {
		return f.FindExtensionsCall.Stub(param1)
	}
	return f.FindExtensionsCall.Returns.String, f.FindExtensionsCall.Returns.Error
}
func (f *FileManager) WriteConfig(param1 string, param2 string, param3 phpdist.PhpIniConfig) (string, string, error) {
	f.WriteConfigCall.mutex.Lock()
	defer f.WriteConfigCall.mutex.Unlock()
	f.WriteConfigCall.CallCount++
	f.WriteConfigCall.Receives.LayerRoot = param1
	f.WriteConfigCall.Receives.CnbPath = param2
	f.WriteConfigCall.Receives.Data = param3
	if f.WriteConfigCall.Stub != nil {
		return f.WriteConfigCall.Stub(param1, param2, param3)
	}
	return f.WriteConfigCall.Returns.DefaultConfig, f.WriteConfigCall.Returns.BuildpackConfig, f.WriteConfigCall.Returns.Err
}
