package iwt

// commit contains the current git commit and is set in the build.sh script
var commit string

// VERSION is the version of this application
var VERSION = "0.1.18" + commit

// APP is the name of the application
const APP string = "IWT Agent"
