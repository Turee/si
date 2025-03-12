package version

// Version information set by build process
var (
	// Version is the current version of the application
	Version = "dev"

	// Commit is the git commit SHA this build was created from
	Commit = "unknown"

	// BuildDate is the date when this build was created
	BuildDate = "unknown"
)

// Info returns a string with version information
func Info() string {
	return "Version: " + Version + " Commit: " + Commit + " BuildDate: " + BuildDate
}
