package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type BackupConfig struct {
	Path      string `mapstructure:"path"`
	Retention int    `mapstructure:"retention"`
	Compress  bool   `mapstructure:"compress"`
}

type ScheduleConfig struct {
	Cron       string `mapstructure:"cron"`
	Timezone   string `mapstructure:"timezone"`
	RunOnStart bool   `mapstructure:"run_on_start"`
}

type AppConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Backup   BackupConfig   `mapstructure:"backup"`
	Schedule ScheduleConfig `mapstructure:"schedule"`
}

var Config AppConfig

func Init(cfgFile string) error {
	// 设置默认值
	viper.SetDefault("database.type", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")

	viper.SetDefault("backup.path", "./backups")
	viper.SetDefault("backup.retention", 30)
	viper.SetDefault("backup.compress", true)

	viper.SetDefault("schedule.cron", "0 0 * * 4") // 每周四00:00
	viper.SetDefault("schedule.timezone", "Local")
	viper.SetDefault("schedule.run_on_start", false)

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// 自动查找配置文件
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(filepath.Join(home, ".dbbackup"))
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/dbbackup")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// 从环境变量读取配置
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DBBACKUP")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在，使用默认值
	} else {
		fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())
	}

	// 反序列化配置
	if err := viper.Unmarshal(&Config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}
