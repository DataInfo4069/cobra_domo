package cmd

//备份命令

import (
	"fmt"
	"time"

	"cobra_domo/config"
	"cobra_domo/internal/backup"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "立即执行数据库备份",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查数据库配置
		if config.Config.Database.Name == "" {
			return fmt.Errorf("数据库名称未配置")
		}

		fmt.Printf("开始备份数据库: %s\n", config.Config.Database.Name)
		fmt.Printf("数据库类型: %s\n", config.Config.Database.Type)
		fmt.Printf("备份位置: %s\n", config.Config.Backup.Path)

		// 执行备份
		result, err := backup.ExecuteBackup(ctx)
		if err != nil {
			return fmt.Errorf("备份失败: %w", err)
		}

		fmt.Printf("备份成功!\n")
		fmt.Printf("文件: %s\n", result.Filepath)
		fmt.Printf("大小: %s\n", backup.FormatFileSize(result.Size))
		fmt.Printf("耗时: %v\n", result.Duration.Round(time.Millisecond))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// 本地标志
	backupCmd.Flags().StringP("output", "o", "", "指定备份文件输出路径")
	backupCmd.Flags().Bool("no-compress", false, "不压缩备份文件")

	// 绑定配置
	viper.BindPFlag("backup.path", backupCmd.Flags().Lookup("output"))
	viper.BindPFlag("backup.compress", backupCmd.Flags().Lookup("no-compress"))
}
