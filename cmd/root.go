package cmd

//根命令

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cobra_domo/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	ctx     context.Context
	cancel  context.CancelFunc
)

var rootCmd = &cobra.Command{
	Use:   "dbbackup",
	Short: "数据库定时备份工具",
	Long: `一个简单高效的数据库定时备份工具，支持 MySQL 和 PostgreSQL。
默认每周四凌晨 12 点自动执行备份。`,
	Version: "1.0.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 初始化配置
		if err := config.Init(cfgFile); err != nil {
			return fmt.Errorf("配置初始化失败: %w", err)
		}

		// 设置上下文，用于优雅关闭
		ctx, cancel = signal.NotifyContext(context.Background(),
			syscall.SIGINT, syscall.SIGTERM)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// 清理资源
		if cancel != nil {
			cancel()
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出模式")

	// 数据库连接标志
	rootCmd.PersistentFlags().String("db-host", "", "数据库主机地址")
	rootCmd.PersistentFlags().Int("db-port", 0, "数据库端口")
	rootCmd.PersistentFlags().String("db-user", "", "数据库用户名")
	rootCmd.PersistentFlags().String("db-password", "", "数据库密码")
	rootCmd.PersistentFlags().String("db-name", "", "数据库名称")
	rootCmd.PersistentFlags().String("db-type", "", "数据库类型 (mysql/postgres)")
}
