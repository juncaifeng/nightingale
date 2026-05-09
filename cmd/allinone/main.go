package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ccfos/nightingale/v6/center"
	"github.com/ccfos/nightingale/v6/pkg/osx"
	"github.com/ccfos/nightingale/v6/pkg/version"
	"github.com/toolkits/pkg/runner"
)

var (
	showVersion = flag.Bool("version", false, "Show version.")
	configDir = flag.String("configs", osx.GetEnv("N9E_CONFIGS", ""), "Specify configuration directory.(env:N9E_CONFIGS)")
	cryptoKey = flag.String("crypto-key", "", "Specify the secret key for configuration file encryption.")
	dataDir = flag.String("data-dir", osx.GetEnv("N9E_DATA_DIR", "./n9e-data"), "Data directory for SQLite and logs.(env:N9E_DATA_DIR)")
	port = flag.Int("port", 0, "HTTP service port.")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	printEnv()

	configPath := *configDir
	if configPath == "" {
		configPath = autoGenerateConfig(*dataDir, *port)
	}

	cleanFunc, err := center.Initialize(configPath, *cryptoKey)
	if err != nil {
		log.Fatalln("failed to initialize:", err)
	}

	code := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

EXIT:
	for {
		sig := <-sc
		fmt.Println("received signal:", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			code = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	fmt.Println("process exited")
	os.Exit(code)
}

func autoGenerateConfig(dataDir string, customPort int) string {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(absDataDir, "logs"), 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	httpPort := 17000
	if customPort > 0 {
		httpPort = customPort
	}

	configContent := fmt.Sprintf(`[Global]
RunMode = "release"

[Log]
Dir = "%s/logs"
Level = "INFO"
Output = "stdout"

[HTTP]
Host = "0.0.0.0"
Port = %d
PrintAccessLog = false
PProf = false
ExposeMetrics = true
ShutdownTimeout = 30
MaxContentLength = 67108864
ReadTimeout = 20
WriteTimeout = 40
IdleTimeout = 120

[HTTP.ShowCaptcha]
Enable = false

[HTTP.APIForAgent]
Enable = true

[HTTP.APIForService]
Enable = false
[HTTP.APIForService.BasicAuth]
user001 = "ccc26da7b9aba533cbb263a36c07dcc5"

[HTTP.JWTAuth]
AccessExpired = 1500
RefreshExpired = 10080
RedisKeyPrefix = "/jwt/"

[HTTP.ProxyAuth]
Enable = false
HeaderUserNameKey = "X-User-Name"
DefaultRoles = ["Standard"]

[HTTP.TokenAuth]
Enable = true

[HTTP.RSA]
OpenRSA = false

[DB]
DBType = "sqlite"
SqliteFile = "%s/n9e.db"
MaxLifetime = 12
MaxOpenConns = 100
MaxIdleConns = 10

[Redis]
Address = "%s"
RedisType = "miniredis"
DB = 0

[Center]
BuiltinIntegrationsDir = ""
OpsYamlFile = ""
MetricsYamlFile = ""
CleanNotifyRecordDay = 365
CleanPipelineExecutionDay = 30
MigrateBusiGroupLabel = false

[Alert]
[Alert.Heartbeat]
IP = "127.0.0.1"
Port = %d
EngineName = "default"

[Alert.Alerting]
[Alert.Alerting.Webhook]
BatchSend = false

[Pushgw]
[Pushgw.Writer]
[Pushgw.Writer.RemoteWrite]
Enabled = true
URL = "http://127.0.0.1:8480/insert/0/prometheus/api/v1/write"
Headers = {}
Timeout = 10
QueueCapacity = 100000

[Ibex]
Enable = false
`, absDataDir+"/logs", httpPort, absDataDir, absDataDir, httpPort)

	configPath := filepath.Join(absDataDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		log.Fatalf("failed to write config file: %v", err)
	}

	fmt.Printf("Configuration auto-generated at: %s\n", configPath)
	fmt.Printf("Data directory: %s\n", absDataDir)
	fmt.Printf("SQLite database: %s/n9e.db\n", absDataDir)
	fmt.Printf("Web interface: http://127.0.0.1:%d\n", httpPort)
	fmt.Println("Default login: root / root.2020")
	fmt.Println()

	return configPath
}

func printEnv() {
	runner.Init()
	fmt.Println("Nightingale All-in-One Version:", version.Version)
	fmt.Println("runner.cwd:", runner.Cwd)
	fmt.Println("runner.hostname:", runner.Hostname)
	fmt.Println("runner.fd_limits:", runner.FdLimits())
	fmt.Println("runner.vm_limits:", runner.VMLimits())
	fmt.Println()
}
