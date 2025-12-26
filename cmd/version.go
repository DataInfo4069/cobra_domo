package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Database Backup Tool v1.0.0")
		fmt.Println("Go 1.25 版本")
		fmt.Println("支持 MySQL 和 PostgreSQL 数据库")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
