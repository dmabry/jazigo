// Package conf handles configuration.
package conf

import (
	"time"

	"gopkg.in/yaml.v3"

	"github.com/udhos/jazigo/store"
)

// Change stores info about last config change.
type Change struct {
	When time.Time
	By   string
	From string
}

// AppConfig is persistent global configuration.
type AppConfig struct {
	MaxConfigFiles    int
	Holdtime          time.Duration
	ScanInterval      time.Duration
	MaxConcurrency    int
	MaxConfigLoadSize int64
	LastChange        Change
	Comment           string // free user-defined field
}

// NewAppConfigFromString creates AppConfig from string.
func NewAppConfigFromString(str string) (*AppConfig, error) {
	b := []byte(str)
	c := &AppConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Dump exports AppConfig as YAML.
func (a *AppConfig) Dump() ([]byte, error) {
	b, err := yaml.Marshal(a)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NewDevAttr creates a new set of DevAttributes.
func NewDevAttr() DevAttributes {
	a := DevAttributes{
		ErrlogHistSize: 60, // default max number of lines in errlog history
	}

	return a
}

// DevAttributes is per-model set of default attributes for device.
type DevAttributes struct {
	NeedLoginChat                bool          // need login chat
	NeedEnabledMode              bool          // need enabled mode
	NeedPagingOff                bool          // need disabled pager
	EnableCommand                string        // enable
	UsernamePromptPattern        string        // Username:
	PasswordPromptPattern        string        // Password:
	EnablePasswordPromptPattern  string        // Password:
	DisabledPromptPattern        string        // >
	EnabledPromptPattern         string        // # ("" --> look for EOF)
	CommandList                  []string      // "show version", "show run"
	DisablePagerCommand          string        // term len 0
	DisablePagerExtraPromptCount int           // consume N extra prompts
	SupressAutoLF                bool          // do not send auto LF
	QuoteSentCommandsFormat      string        // !![%s] - empty means omitting
	KeepControlChars             bool          // enable if you want to capture control chars (backspace, etc)
	LineFilter                   string        // line filter name - applied to every saved line
	ChangesOnly                  bool          // save new file only if it differs from previous one
	S3ContentType                string        // ""=none "detect"=http.Detect "text/plain" etc
	RunProg                      []string      // "/path/to/external/command", "arg1", "arg2" for the run model
	RunTimeout                   time.Duration // 60s - time allowed for external program to complete
	ErrlogHistSize               int           // max number of lines in errlog history
	PostLoginPromptPattern       string        // mikrotik: Please press "Enter" to continue!
	PostLoginPromptResponse      string        // mikrotik: \r\n
	UsernameAppend               string        // mikrotik: +cte

	// readTimeout: per-read timeout (protection against inactivity)
	// matchTimeout: full match timeout (protection against slow sender -- think 1 byte per second)
	ReadTimeout         time.Duration // protection against inactivity
	MatchTimeout        time.Duration // protection against slow sender
	SendTimeout         time.Duration // protection against inactivity
	CommandReadTimeout  time.Duration // larger timeout for slow responses (slow show running)
	CommandMatchTimeout time.Duration // larger timeout for slow responses (slow show running)
}

// DevConfig is full set of device properties.
type DevConfig struct {
	Debug          bool
	Deleted        bool
	Model          string
	ID             string
	HostPort       string
	Transports     string
	LoginUser      string
	LoginPassword  string
	EnablePassword string
	Comment        string // free user-defined field
	LastChange     Change
	Attr           DevAttributes
}

// NewDeviceFromString creates device configuration from string.
func NewDeviceFromString(str string) (*DevConfig, error) {
	b := []byte(str)
	c := &DevConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Dump exports device properties as YAML.
func (c *DevConfig) Dump() ([]byte, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Config is full (global+devices) app configuration.
type Config struct {
	Options AppConfig
	Devices []DevConfig
}

// New creates new full app configuration.
func New() *Config {
	return &Config{
		Options: AppConfig{
			Holdtime:          12 * time.Hour,   // do not retry a successful device backup before this holdtime
			ScanInterval:      10 * time.Minute, // interval for scanning device table
			MaxConcurrency:    20,               // limit for concurrent backup jobs
			MaxConfigFiles:    120,              // limit for per-device saved files
			MaxConfigLoadSize: 10000000,         // 10M limit max config file size for loading to memory
		},
		Devices: []DevConfig{},
	}
}

// Load loads a Config from file.
func Load(path string, maxSize int64) (*Config, error) {
	b, readErr := store.FileRead(path, maxSize)
	if readErr != nil {
		return nil, readErr
	}
	c := New()
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Dump exports device properties as YAML.
func (c *Config) Dump() ([]byte, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return b, nil
}
