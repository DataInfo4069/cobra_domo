package cmd

//定时调度命令

import (
	"fmt"
	"log/slog"
	"os"
	
	"cobra_domo/config"
	"cobra_domo/internal/backup"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "启动定时备份服务",
	Long: `启动定时备份服务，根据配置的 Cron 表达式定期执行备份。
默认：每周四凌晨 12 点 (0 0 * * 4)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 设置日志
		if verbose {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}

		// 获取 Cron 表达式
		cronExpr := config.Config.Schedule.Cron
		if cronExpr == "" {
			cronExpr = "0 0 * * 4" // 每周四00:00
		}

		// 验证 Cron 表达式
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour |
			cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(cronExpr); err != nil {
			return fmt.Errorf("无效的 Cron 表达式 '%s': %w", cronExpr, err)
		}

		fmt.Println("数据库定时备份服务启动")
		fmt.Printf("定时规则: %s\n", cronExpr)
		fmt.Printf("数据库: %s (%s)\n",
			config.Config.Database.Name,
			config.Config.Database.Type)
		fmt.Printf("备份路径: %s\n", config.Config.Backup.Path)
		fmt.Printf("保留天数: %d\n", config.Config.Backup.Retention)
		fmt.Println("服务运行中... (Ctrl+C 退出)")

		// 创建调度器
		c := cron.New(cron.WithParser(parser))

		// 添加定时任务
		jobID, err := c.AddFunc(cronExpr, func() {
			slog.Info("定时备份任务开始执行")

			result, err := backup.ExecuteBackup(ctx)
			if err != nil {
				slog.Error("定时备份失败", "error", err)
				return
			}

			slog.Info("定时备份完成",
				"file", result.Filepath,
				"size", backup.FormatFileSize(result.Size),
				"duration", result.Duration)
		})

		if err != nil {
			return fmt.Errorf("添加定时任务失败: %w", err)
		}

		// 启动时立即执行一次（如果配置了）
		if config.Config.Schedule.RunOnStart {
			fmt.Println("🔧 启动时立即执行备份...")
			go func() {
				if _, err := backup.ExecuteBackup(ctx); err != nil {
					slog.Error("启动时备份失败", "error", err)
				}
			}()
		}

		// 显示下次执行时间
		entry := c.Entry(jobID)
		fmt.Printf("下次执行时间: %s\n",
			entry.Next.Format("2006-01-02 15:04:05"))

		// 启动调度器
		c.Start()

		// 等待上下文取消（优雅关闭）
		<-ctx.Done()

		fmt.Println("\n接收到关闭信号，正在停止调度器...")

		// 停止调度器
		stopCtx := c.Stop()
		<-stopCtx.Done()

		fmt.Println("调度器已安全停止")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmd)

	// 调度器相关标志
	scheduleCmd.Flags().String("cron", "", "Cron 表达式 (默认: 0 0 * * 4)")
	scheduleCmd.Flags().String("timezone", "", "时区设置")
	scheduleCmd.Flags().Bool("run-on-start", false, "启动时立即执行一次")

	// 绑定配置
	viper.BindPFlag("schedule.cron", scheduleCmd.Flags().Lookup("cron"))
	viper.BindPFlag("schedule.timezone", scheduleCmd.Flags().Lookup("timezone"))
	viper.BindPFlag("schedule.run_on_start", scheduleCmd.Flags().Lookup("run-on-start"))
}
