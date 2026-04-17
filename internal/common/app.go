package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/pflag"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/config"
)

// App holds shared application state including the config and API client.
// For CLI commands, it is populated once in PersistentPreRunE and shared across
// all command handlers. For MCP tool handlers, Load is called on each tool
// invocation to refresh config and credentials (in case concurrent CLI
// commands have changed config options or logged the user in/out while the MCP
// session was ongoing).
//
// All fields are unexported. Use Load or SetClient to populate state, and
// GetAll/GetConfig/GetClient to read it. Concurrency is handled internally
// via a sync.RWMutex.
// ClientFactory creates an API client from the loaded config.
// Used in tests to inject a mock client while still allowing
// PersistentPreRunE to run normally (loading config, etc.).
type ClientFactory func(ctx context.Context, cfg *config.Config) (api.ClientWithResponsesInterface, string, error)

type App struct {
	Experimental  bool
	flags         *pflag.FlagSet
	config        *config.Config
	client        api.ClientWithResponsesInterface // nil if credentials unavailable
	projectID     string
	clientErr     error         // returned by GetClient() when client is nil
	clientFactory ClientFactory // nil in production; set in tests
	lock          sync.RWMutex  // protects config, client, projectID, clientErr
}

// SetFlags stores the command's flag set for use by config.Load. Must be
// called before Load (typically in PersistentPreRunE).
func (a *App) SetFlags(flags *pflag.FlagSet) {
	a.flags = flags
}

// SetClientFactory sets a custom factory for API client creation.
// When set, Load() calls this instead of the default newAPIClient.
func (a *App) SetClientFactory(f ClientFactory) {
	a.clientFactory = f
}

// Load loads (or reloads) the config from disk and attempts to create the API
// client. Returns the loaded config, API client, and project ID. Config
// loading errors are returned; API client errors are stored internally and
// surfaced via GetClient (the returned client will simply be nil if it
// couldn't be loaded).
func (a *App) Load(ctx context.Context) (*config.Config, api.ClientWithResponsesInterface, string, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	cfg, err := config.Load(a.flags)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to load config: %w", err)
	}
	a.config = cfg

	a.client, a.projectID, a.clientErr = a.newAPIClient(ctx, a.config)

	return a.config, a.client, a.projectID, nil
}

// SetClient stores an existing API client and project ID on the App. Use this
// when a valid client already exists (e.g. after login creates one for
// validation) to avoid redundantly re-reading credentials from disk.
func (a *App) SetClient(client api.ClientWithResponsesInterface, projectID string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.client = client
	a.projectID = projectID
	a.clientErr = nil
}

func (a *App) newAPIClient(ctx context.Context, cfg *config.Config) (api.ClientWithResponsesInterface, string, error) {
	if a.clientFactory != nil {
		return a.clientFactory(ctx, a.config)
	}
	return newAPIClient(ctx, cfg)
}

// TryGetAll returns a snapshot of the config, API client, and project ID. The
// returned values remain valid even if Load is called concurrently — the
// pointers reference objects that are not mutated by Load (Load replaces
// pointers, it does not modify the objects they point to). Does not return an
// error if the client is not available (e.g. user is not logged in).
func (a *App) TryGetAll() (*config.Config, api.ClientWithResponsesInterface, string) {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.config, a.client, a.projectID
}

// GetAll returns a snapshot of the config, API client, and project ID. The
// returned values remain valid even if Load is called concurrently — the
// pointers reference objects that are not mutated by Load (Load replaces
// pointers, it does not modify the objects they point to). Returns an
// error if the client is not available (e.g. user is not logged in).
func (a *App) GetAll() (*config.Config, api.ClientWithResponsesInterface, string, error) {
	cfg, client, projectID := a.TryGetAll()
	if a.client == nil {
		return nil, nil, "", a.clientErr
	}
	return cfg, client, projectID, nil
}

// GetConfig returns a snapshot of the config.
func (a *App) GetConfig() *config.Config {
	cfg, _, _ := a.TryGetAll()
	return cfg
}

// GetClient returns a snapshot of the API client and project ID. Returns an
// error if the client is not available (e.g. user is not logged in).
func (a *App) GetClient() (api.ClientWithResponsesInterface, string, error) {
	_, client, projectID, err := a.GetAll()
	return client, projectID, err
}
