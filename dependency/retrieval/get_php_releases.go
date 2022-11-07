package retrieval

import "github.com/joshuatcasey/libdependency/upstream"

//go:generate faux --interface GetPhpReleases --output fakes/get_php_releases.go
type GetPhpReleases interface {
	Get(string) (PhpReleaseAnouncementsRaw, error)
}

type GetPhpReleasesImpl struct{}

func (i GetPhpReleasesImpl) Get(indexUrl string) (PhpReleaseAnouncementsRaw, error) {
	var indexData PhpReleaseAnouncementsRaw

	err := upstream.GetAndUnmarshal(indexUrl, &indexData)
	return indexData, err
}

type PhpReleaseAnouncementsRaw map[string]PhpReleaseLineAnnouncementsRaw
type PhpReleaseLineAnnouncementsRaw map[string]PhpReleaseAnnouncementRaw
type PhpReleaseAnnouncementRaw struct {
	Date    string                `json:"date"`
	Source  []PhpReleaseSourceRaw `json:"source"`
	Version string                `json:"version"`
}
type PhpReleaseSourceRaw struct {
	Filename string `json:"filename"`
	Sha256   string `json:"sha256"`
}
