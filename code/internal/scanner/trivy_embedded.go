package scanner

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// 嵌入Trivy二进制文件（CI/CD构建时按目标平台下载）
//go:embed binaries/*
var embeddedBinaries embed.FS

// TrivyBinaryManager 管理Trivy二进制文件的提取和使用
type TrivyBinaryManager struct {
	extractedPath string
	version       string
	extracted     bool
}

// NewTrivyBinaryManager 创建新的Trivy二进制管理器
func NewTrivyBinaryManager() *TrivyBinaryManager {
	return &TrivyBinaryManager{
		version: "0.48.3",
	}
}

// ExtractBinary 提取当前平台对应的Trivy二进制文件
func (tbm *TrivyBinaryManager) ExtractBinary() error {
	if tbm.extracted && tbm.isBinaryValid() {
		return nil // 已提取且有效
	}

	// 1. 首先尝试使用本地文件
	if localPath := tbm.getLocalBinaryPath(); localPath != "" && tbm.isLocalBinaryValid(localPath) {
		tbm.extractedPath = localPath
		tbm.extracted = true
		return nil
	}

	// 2. 回退到嵌入文件提取逻辑
	// 确定当前平台的二进制文件名
	embeddedFileName, extractedFileName, err := tbm.getPlatformBinaryNames()
	if err != nil {
		return err
	}

	// 从嵌入文件系统中读取二进制数据
	binaryData, err := embeddedBinaries.ReadFile(embeddedFileName)
	if err != nil {
		return fmt.Errorf("读取嵌入的二进制文件失败 (%s): %w", embeddedFileName, err)
	}

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "inspector-binaries")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 写入到临时文件
	tbm.extractedPath = filepath.Join(tempDir, extractedFileName)
	if err := os.WriteFile(tbm.extractedPath, binaryData, 0755); err != nil {
		return fmt.Errorf("写入二进制文件失败: %w", err)
	}

	tbm.extracted = true
	return nil
}

// GetBinaryPath 获取提取的二进制文件路径
func (tbm *TrivyBinaryManager) GetBinaryPath() (string, error) {
	if !tbm.extracted {
		if err := tbm.ExtractBinary(); err != nil {
			return "", err
		}
	}
	return tbm.extractedPath, nil
}

// IsAvailable 检查Trivy二进制是否可用
func (tbm *TrivyBinaryManager) IsAvailable() bool {
	if err := tbm.ExtractBinary(); err != nil {
		return false
	}
	return tbm.isBinaryValid()
}

// GetVersion 获取Trivy版本
func (tbm *TrivyBinaryManager) GetVersion() string {
	return tbm.version
}

// Cleanup 清理临时文件
func (tbm *TrivyBinaryManager) Cleanup() error {
	if tbm.extractedPath != "" {
		return os.Remove(tbm.extractedPath)
	}
	return nil
}

// getPlatformBinaryNames 根据当前平台确定二进制文件名
func (tbm *TrivyBinaryManager) getPlatformBinaryNames() (embeddedFileName, extractedFileName string, err error) {
	switch runtime.GOOS {
	case "linux":
		return "binaries/trivy-linux-amd64", "trivy", nil
	case "darwin":
		return "binaries/trivy-darwin-amd64", "trivy", nil
	case "windows":
		return "binaries/trivy-windows-amd64.exe", "trivy.exe", nil
	default:
		return "", "", fmt.Errorf("不支持的操作系统: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

// isBinaryValid 检查已提取的二进制文件是否有效
func (tbm *TrivyBinaryManager) isBinaryValid() bool {
	if tbm.extractedPath == "" {
		return false
	}

	// 检查文件是否存在
	if _, err := os.Stat(tbm.extractedPath); os.IsNotExist(err) {
		return false
	}

	// 可以进一步检查文件权限和版本
	// 这里简化处理，只检查文件存在性
	return true
}

// GetSupportedPlatforms 获取支持的平台列表
func GetSupportedPlatforms() []string {
	return []string{
		"linux/amd64",
		"darwin/amd64", 
		"windows/amd64",
	}
}

// GetCurrentPlatform 获取当前平台标识
func GetCurrentPlatform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// getLocalBinaryPath 获取本地二进制文件路径
func (tbm *TrivyBinaryManager) getLocalBinaryPath() string {
	// 确定当前平台的本地二进制文件名（使用实际的本地文件名格式）
	var localFileName string
	switch runtime.GOOS {
	case "linux":
		localFileName = "trivy-linux-amd64"
	case "darwin":
		localFileName = "trivy-darwin-amd64"
	case "windows":
		localFileName = "trivy-windows-amd64.exe"
	default:
		return ""
	}

	// 构建本地文件路径，支持多种可能的工作目录
	possiblePaths := []string{
		// 从项目根目录执行
		filepath.Join("code", "internal", "scanner", "binaries", localFileName),
		// 从 code 目录执行
		filepath.Join("internal", "scanner", "binaries", localFileName),
		// 从当前目录执行（如果就在 scanner 目录）
		filepath.Join("binaries", localFileName),
		// 绝对路径（如果设置了环境变量等）
		filepath.Join(".", "code", "internal", "scanner", "binaries", localFileName),
	}

	// 检查每个可能的路径
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			// 直接返回相对路径，避免绝对路径可能的性能问题
			return path
		}
	}

	return ""
}

// isLocalBinaryValid 检查本地二进制文件是否有效
func (tbm *TrivyBinaryManager) isLocalBinaryValid(path string) bool {
	if path == "" {
		return false
	}

	// 检查文件是否存在
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	// 检查是否是普通文件
	if !info.Mode().IsRegular() {
		return false
	}

	// 检查文件大小（Trivy二进制文件应该比较大）
	if info.Size() < 1024*1024 { // 小于1MB可能不是有效的Trivy二进制
		return false
	}

	// 在Windows上检查是否有执行权限
	if runtime.GOOS == "windows" {
		return filepath.Ext(path) == ".exe"
	}

	// 在Unix系统上检查执行权限
	return info.Mode()&0111 != 0
}
