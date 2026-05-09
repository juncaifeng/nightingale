package center

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ccfos/nightingale/v6/conf"
)

type AllInOneConfig struct {
	Port     int
	DataDir  string
	Password string
}

func GenerateAllInOneConfig(dataDir string) (string, error) {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
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
`, absDataDir+"/logs", 17000, absDataDir, absDataDir, 17000)

	configPath := filepath.Join(absDataDir, "config.toml")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

func InitAllInOne(configDir string) (*conf.ConfigType, error) {
	if configDir == "" {
		configDir = "./etc"
	}

	config, err := conf.InitConfig(configDir, "")
	if err != nil {
		return nil, fmt.Errorf("failed to init config: %v", err)
	}

	return config, nil
}
