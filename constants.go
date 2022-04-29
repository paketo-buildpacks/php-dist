package phpdist

const (
	DepKey           = "dependency-sha"
	PHPDependency    = "php"
	UserProvidedPath = ".php.ini.d"
)

var EntryPriorities = []interface{}{
	"BP_PHP_VERSION",
	"composer.lock",
	"composer.json",
	"default-versions",
}
