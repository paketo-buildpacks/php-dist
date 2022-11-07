package fakes

import (
	"sync"

	"github.com/paketo-buildpacks/php-dist/retrieval"
)

type GetPhpReleases struct {
	GetCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			String string
		}
		Returns struct {
			PhpReleaseAnouncementsRaw retrieval.PhpReleaseAnouncementsRaw
			Error                     error
		}
		Stub func(string) (retrieval.PhpReleaseAnouncementsRaw, error)
	}
}

func (f *GetPhpReleases) Get(param1 string) (retrieval.PhpReleaseAnouncementsRaw, error) {
	f.GetCall.mutex.Lock()
	defer f.GetCall.mutex.Unlock()
	f.GetCall.CallCount++
	f.GetCall.Receives.String = param1
	if f.GetCall.Stub != nil {
		return f.GetCall.Stub(param1)
	}
	return f.GetCall.Returns.PhpReleaseAnouncementsRaw, f.GetCall.Returns.Error
}
