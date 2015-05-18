package devastator

import (
	"log"
	"os"
)

const (
	// devastator server envrinment variables
	goEnv           = "GO_ENV"
	devastatorEnv   = "DEVASTATOR_ENV"
	devastatorDebug = "DEVASTATOR_DEBUG"
	devastatorPort  = "DEVASTATOR_PORT"

	// possible DEVASTATOR_ENV values
	envDev  = "development"
	envTest = "test"
	envProd = "production"

	// GCM environment variables
	gcmSenderID = "GCM_SENDER_ID"
	gcmCcsHost  = "GCM_CCS_HOST"

	// Google environment variables
	googleAPIKey = "GOOGLE_API_KEY"

	// Default listener port configuration
	portDefault = "3000"
	portTest    = "3001"
)

// Conf contains all the global configuration for the devastator server.
var Conf Config

// Config describes the global configuration for the devastator server.
type Config struct {
	App App
	GCM GCM
}

// App contains the global application variables.
type App struct {
	// One of the following: development, test, production
	Env string
	// Enables verbose logging to stdout
	Debug bool
	// Listener port
	Port string
}

// GCM describes the Google Cloud Messaging parameters as described here: https://developer.android.com/google/gcm/gs.html
type GCM struct {
	CCSHost  string
	SenderID string
}

// APIKey gets the GCM API key from environment variable.
func (gcm *GCM) APIKey() string {
	return os.Getenv(googleAPIKey)
}

// InitConf initializes application configuration.
// If given, env parameter overrides environment configuration. This is useful for testing.
func InitConf(env string) {
	if env == "" {
		env = os.Getenv(devastatorEnv)
	}
	if env == "" {
		if env = os.Getenv(goEnv); env == "" {
			env = envDev
		}
	}
	debug := os.Getenv(devastatorDebug) != "" || (env != envProd)
	port := os.Getenv(devastatorPort)
	if port == "" {
		switch env {
		case envTest:
			port = portTest
		default:
			port = portDefault
		}
	}

	app := App{Env: env, Debug: debug, Port: port}
	gcm := GCM{CCSHost: os.Getenv(gcmCcsHost), SenderID: os.Getenv(gcmSenderID)}
	Conf = Config{App: app, GCM: gcm}
	log.Printf("Server config initialized with values: %+v\n", Conf)
}
