module github.com/paketo-buildpacks/php-dist/retrieval

go 1.22

// This is required because of a breaking change in a newer version
replace github.com/ekzhu/minhash-lsh => github.com/ekzhu/minhash-lsh v0.0.0-20171225071031-5c06ee8586a1

require (
	github.com/Masterminds/semver/v3 v3.2.1
	github.com/joshuatcasey/libdependency v0.25.0
	github.com/paketo-buildpacks/packit/v2 v2.14.0
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/anchore/packageurl-go v0.1.1-0.20230104203445-02e0a6721501 // indirect
	github.com/cloudflare/circl v1.3.8 // indirect
	github.com/cyphar/filepath-securejoin v0.2.5 // indirect
	github.com/dgryski/go-minhash v0.0.0-20190315135803-ad340ca03076 // indirect
	github.com/ekzhu/minhash-lsh v0.0.0-20190924033628-faac2c6342f8 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-enry/go-license-detector/v4 v4.3.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-git/go-git/v5 v5.12.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/hhatto/gorst v0.0.0-20181029133204-ca9f730cac5b // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jdkato/prose v1.2.1 // indirect
	github.com/joshuatcasey/collections v0.5.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shogo82148/go-shuffle v1.0.1 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/ulikunitz/xz v0.5.12 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gonum.org/v1/gonum v0.15.0 // indirect
	gopkg.in/neurosnap/sentences.v1 v1.0.7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)
