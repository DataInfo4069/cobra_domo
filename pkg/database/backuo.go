package database

//数据库逻辑备份

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DBConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// Backup 执行数据库备份
func Backup(config DBConfig, backupPath string) (string, error) {
	// 确保备份目录存在
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("创建备份目录失败: %v", err)
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s.sql.gz", config.Name, timestamp)
	filepath := filepath.Join(backupPath, filename)

	var cmd *exec.Cmd

	switch strings.ToLower(config.Type) {
	case "mysql", "mariadb":
		cmd = exec.Command("mysqldump",
			"-h", config.Host,
			"-P", fmt.Sprintf("%d", config.Port),
			"-u", config.User,
			fmt.Sprintf("-p%s", config.Password),
			config.Name,
		)
	case "postgres", "postgresql":
		os.Setenv("PGPASSWORD", config.Password)
		cmd = exec.Command("pg_dump",
			"-h", config.Host,
			"-p", fmt.Sprintf("%d", config.Port),
			"-U", config.User,
			"-d", config.Name,
		)
	default:
		return "", fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	// 创建压缩命令
	gzipCmd := exec.Command("gzip")

	// 设置管道: mysqldump -> gzip -> 文件
	gzipCmd.Stdin, _ = cmd.StdoutPipe()
	output, _ := os.Create(filepath)
	defer output.Close()
	gzipCmd.Stdout = output

	// 启动命令
	if err := gzipCmd.Start(); err != nil {
		return "", fmt.Errorf("启动gzip失败: %v", err)
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("数据库导出失败: %v", err)
	}

	if err := gzipCmd.Wait(); err != nil {
		return "", fmt.Errorf("压缩失败: %v", err)
	}

	return filepath, nil
}
func CleanOldBackups(backupPath string, keepDays int) int {
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return 0
	}

	// 获取文件信息并过滤备份文件
	var backupFiles []os.FileInfo
	for _, entry := range entries {
		if info, err := entry.Info(); err == nil {
			if strings.HasPrefix(info.Name(), "backup_") && strings.HasSuffix(info.Name(), ".sql.gz") {
				backupFiles = append(backupFiles, info)
			}
		}
	}

	// 按修改时间排序（从旧到新）
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].ModTime().Before(backupFiles[j].ModTime())
	})

	// 计算保留截止时间
	cutoffTime := time.Now().AddDate(0, 0, -keepDays)

	// 删除旧文件
	deleted := 0
	for _, f := range backupFiles {
		if f.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(backupPath, f.Name())
			if err := os.Remove(filePath); err == nil {
				deleted++
			}
		}
	}

	return deleted
}
