package phpdist

const (
	DepKey        = "dependency-sha"
	PHPDependency = "php"
)

var EntryPriorities = []interface{}{
	"BP_PHP_VERSION",
	"buildpack.yml",
	"composer.lock",
	"composer.json",
	"default-versions",
}
