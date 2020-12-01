package evo

import (
	"fmt"
	"github.com/getevo/evo/lib/gpath"
	"github.com/getevo/evo/lib/text"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var WorkingDir string
var Exec string
var OS = struct {
	Name     string
	Version  string
	Kernel   string
	Username string
	Path     struct {
		Bin     []string
		Home    string
		Root    string
		Tmp     string
		AppData []string
	}
}{}
var ProcessID int

var DefaultPath = map[string][]string{}

type Configuration struct {
	Log struct {
		WriteFile bool   `yaml:"writefile"`
		Path      string `yaml:"path"`
		Level     string `yaml:"level"`
	} `yaml:"log"`

	App struct {
		Name       string `yaml:"name"`
		Static     string `yaml:"static"`
		SessionAge int    `yaml:"session-age"`
		Language   string `yaml:"language"`
		StrongPass int    `yaml:"strong-pass-level"`
	} `yaml:"app"`

	JWT struct {
		Secret    string        `yaml:"secret"`
		Issuer    string        `yaml:"issuer"`
		Audience  []string      `yaml:"audience"`
		Age       time.Duration `yaml:"age"`
		Subject   string        `yaml:"subject"`
		AgeString string        `yaml:"age"`
	} `yaml:"jwt"`

	Server struct {
		Host          string `yaml:"host"`
		Port          string `yaml:"port"`
		Cert          string `yaml:"cert"`
		Key           string `yaml:"key"`
		HTTPS         bool   `yaml:"https"`
		Name          string `yaml:"name"`
		MaxUploadSize string `yaml:"max-upload-size"`
		StrictRouting bool   `yaml:"strict-routing"`
		CaseSensitive bool   `yaml:"case-sensitive"`
		RequestID     bool   `yaml:"request-id"`
		Debug         bool   `yaml:"debug"`
		Recover       bool   `yaml:"recover"`
	} `yaml:"server"`

	Database struct {
		Enabled        bool          `yaml:"enabled"`
		Type           string        `yaml:"type"`
		Username       string        `yaml:"user"`
		Password       string        `yaml:"pass"`
		Server         string        `yaml:"server"`
		Cache          string        `yaml:"cache"`
		CacheSize      string        `yaml:"cache-size"`
		Debug          string        `yaml:"debug"`
		Database       string        `yaml:"database"`
		SSLMode        string        `yaml:"ssl-mode"`
		Params         string        `yaml:"params"`
		MaxOpenConns   int           `yaml:"max-open-connections"`
		MaxIdleConns   int           `yaml:"max-idle-connections"`
		ConnMaxLifTime time.Duration `yaml:"connection-max-lifetime"`
	} `yaml:"database"`

	Tweaks struct {
		Ballast       bool   `yaml:"ballast"`
		BallastSize   string `yaml:"ballast-size"`
		MaxProcessors int    `yaml:"processors"`
		PreFork       bool   `yaml:"prefork"`
	} `yaml:"tweaks"`

	RateLimit struct {
		Enabled  bool `yaml:"enabled"`
		Duration int  `yaml:"duration"`
		Requests int  `yaml:"requests"`
	} `yaml:"ratelimit"`

	CORS struct {
		Enabled          bool     `yaml:"enabled"`
		AllowOrigins     string   `yaml:"allowed-origins"`
		AllowMethods     []string `yaml:"allowed-methods"`
		AllowHeaders     []string `yaml:"allowed-headers"`
		AllowCredentials bool     `yaml:"allowed-credentials"`
		MaxAge           int      `yaml:"requests"`
	} `yaml:"cors"`
}

var config = &Configuration{}

func parseConfig() *Configuration {
	WorkingDir, _ = os.Getwd()
	WorkingDir = gpath.RSlash(WorkingDir)
	OS.Name = runtime.GOOS
	if OS.Name == "windows" {
		Exec = strings.Trim(filepath.Base(os.Args[0]), ".exe")
		OS.Username = echo("%USERNAME%")
		OS.Path.Bin = []string{}
		for _, item := range strings.Split(echo("%PATH%"), ";") {
			if item != "" {
				OS.Path.Bin = append(OS.Path.Bin, item)
			}
		}
		OS.Path.Home = echo("%USERPROFILE%")
		OS.Path.AppData = []string{WorkingDir, echo("%USERPROFILE%"), echo("%LOCALAPPDATA%"), echo("%HOMEPATH%"), echo("%APPDATA%"), echo("%ALLUSERSPROFILE%")}
		OS.Path.Root = echo("%SystemRoot%")
		OS.Path.Tmp = echo("%TMP%")
		for i, item := range OS.Path.AppData {
			if gpath.IsDirExist(gpath.RSlash(item) + "\\" + Exec) {
				OS.Path.AppData[i] = gpath.RSlash(item) + "\\" + Exec
			}
		}
		OS.Kernel = text.ParseWildCard(run("ver"), `\[Version *\]`)[0]
		OS.Version = OS.Kernel
	} else {
		Exec = filepath.Base(os.Args[0])
		OS.Username = run("whoami")
		OS.Path.Bin = []string{}
		for _, item := range strings.Split(os.Getenv("PATH"), ":") {
			if item != "" {
				OS.Path.Bin = append(OS.Path.Bin, item)
			}
		}
		OS.Path.Home = os.Getenv("HOME")
		OS.Path.AppData = []string{WorkingDir, "/", "/var", "/etc", OS.Path.Home}
		OS.Path.Root = echo("/")
		OS.Path.Tmp = echo("/tmp")
		for i, item := range OS.Path.AppData {
			if gpath.IsDirExist(gpath.RSlash(item) + "\\" + Exec) {
				OS.Path.AppData[i] = gpath.RSlash(item) + "\\" + Exec
			}
		}

		OS.Kernel = run("uname -r")
		OS.Version = text.ParseWildCard(run("cat /etc/os-release | grep PRETTY_NAME"), "\"*\"")[0]

	}
	m := map[string]interface{}{}

	data, err := ioutil.ReadFile(GuessPath(Arg.Config))
	if err != nil {
		log.Println("could not load config file at %s", GuessPath(Arg.Config))
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Println("config syntax error ", GuessPath(Arg.Config))
		log.Fatalf("error: %v", err)
	}

	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &config,
		TagName:  "yaml",
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		log.Println("config reader error ", GuessPath(Arg.Config))
		log.Fatalf("error: %v", err)
	}
	decoder.Decode(m)

	config.App.Static = gpath.RSlash(config.App.Static)

	//yaml string to time.duration bug
	/*s2dParser := str2duration.NewStr2DurationParser()
	age, err := s2dParser.Str2Duration(config.JWT.AgeString)
	if err == nil {
		config.JWT.Age = age
	}*/
	return config
}

// Guess path find file path from different locations
func GuessPath(file string) string {
	wd, _ := os.Getwd()
	if gpath.IsFileExist(wd + "/" + file) {
		return wd + "/" + file
	}
	if gpath.IsFileExist(file) {
		return file
	}
	if paths, has := DefaultPath[file]; has {
		for _, item := range paths {
			if gpath.IsFileExist(gpath.RSlash(item) + "/" + file) {
				return gpath.RSlash(item) + "/" + file
			}
		}
	}

	for _, item := range OS.Path.AppData {
		if gpath.IsFileExist(gpath.RSlash(item) + "/" + file) {
			return gpath.RSlash(item) + "/" + file
		}
	}
	log.Println("unable to guess file path", file)
	return file
}

func echo(command string) string {
	var b []byte
	if runtime.GOOS == "windows" {
		b, _ = exec.Command("cmd", "/C", "echo", command).CombinedOutput()
	} else {
		b, _ = exec.Command("echo", command).CombinedOutput()
	}
	return strings.TrimSpace(string(b))
}

func run(command string) string {
	var b []byte
	if runtime.GOOS == "windows" {
		b, _ = exec.Command("cmd", "/C", command).CombinedOutput()
	} else {
		b, _ = exec.Command("bash", "-c", command).CombinedOutput()
	}
	return strings.TrimSpace(string(b))
}

// GetConfig return configuration instance
func GetConfig() *Configuration {
	return config
}

func ParseConfig(path string, key string, out interface{}) error {
	if path == "" {
		path = Arg.Config
	}
	path = GuessPath(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not load config file at %s", path)
	}

	m := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return err
	}
	var toDecode interface{}
	if key != "" {
		if v, ok := m[key]; ok {
			toDecode = v
		} else {
			return fmt.Errorf("cannot find %s in %s", key, path)
		}
	} else {
		toDecode = m
	}

	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &out,
		TagName:  "yaml",
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	err = decoder.Decode(toDecode)
	if err != nil {
		return err
	}
	return nil
}
