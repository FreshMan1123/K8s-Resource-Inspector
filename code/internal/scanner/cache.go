package scanner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VulnerabilityCache 漏洞扫描结果缓存
type VulnerabilityCache struct {
	cacheDir string
	ttl      time.Duration
	enabled  bool
}

// NewVulnerabilityCache 创建新的漏洞缓存
func NewVulnerabilityCache(cacheDir string, ttl time.Duration) *VulnerabilityCache {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "inspector-vuln-cache")
	}
	
	return &VulnerabilityCache{
		cacheDir: cacheDir,
		ttl:      ttl,
		enabled:  true,
	}
}

// Get 从缓存中获取漏洞报告
func (vc *VulnerabilityCache) Get(key string) (*VulnerabilityReport, bool) {
	if !vc.enabled {
		return nil, false
	}

	cacheFile := vc.getCacheFile(key)
	
	// 检查文件是否存在
	info, err := os.Stat(cacheFile)
	if err != nil {
		return nil, false
	}
	
	// 检查是否过期
	if time.Since(info.ModTime()) > vc.ttl {
		os.Remove(cacheFile) // 删除过期文件
		return nil, false
	}
	
	// 读取缓存文件
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}
	
	// 反序列化
	var report VulnerabilityReport
	if err := json.Unmarshal(data, &report); err != nil {
		os.Remove(cacheFile) // 删除损坏的缓存文件
		return nil, false
	}
	
	return &report, true
}

// Set 将漏洞报告存储到缓存
func (vc *VulnerabilityCache) Set(key string, report *VulnerabilityReport) error {
	if !vc.enabled {
		return nil
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(vc.cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}
	
	// 序列化报告
	data, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("序列化报告失败: %w", err)
	}
	
	// 写入缓存文件
	cacheFile := vc.getCacheFile(key)
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}
	
	return nil
}

// Delete 删除指定的缓存项
func (vc *VulnerabilityCache) Delete(key string) error {
	if !vc.enabled {
		return nil
	}

	cacheFile := vc.getCacheFile(key)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除缓存文件失败: %w", err)
	}
	
	return nil
}

// Clear 清空所有缓存
func (vc *VulnerabilityCache) Clear() error {
	if !vc.enabled {
		return nil
	}

	if err := os.RemoveAll(vc.cacheDir); err != nil {
		return fmt.Errorf("清空缓存目录失败: %w", err)
	}
	
	return nil
}

// IsEnabled 检查缓存是否启用
func (vc *VulnerabilityCache) IsEnabled() bool {
	return vc.enabled
}

// SetEnabled 设置缓存启用状态
func (vc *VulnerabilityCache) SetEnabled(enabled bool) {
	vc.enabled = enabled
}

// GetStats 获取缓存统计信息
func (vc *VulnerabilityCache) GetStats() (CacheStats, error) {
	stats := CacheStats{
		CacheDir: vc.cacheDir,
		TTL:      vc.ttl,
		Enabled:  vc.enabled,
	}
	
	if !vc.enabled {
		return stats, nil
	}
	
	// 统计缓存文件
	entries, err := os.ReadDir(vc.cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return stats, nil // 缓存目录不存在，返回空统计
		}
		return stats, fmt.Errorf("读取缓存目录失败: %w", err)
	}
	
	var totalSize int64
	validCount := 0
	expiredCount := 0
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		totalSize += info.Size()
		
		// 检查是否过期
		if time.Since(info.ModTime()) > vc.ttl {
			expiredCount++
		} else {
			validCount++
		}
	}
	
	stats.TotalFiles = len(entries)
	stats.ValidFiles = validCount
	stats.ExpiredFiles = expiredCount
	stats.TotalSize = totalSize
	
	return stats, nil
}

// getCacheFile 根据键生成缓存文件路径
func (vc *VulnerabilityCache) getCacheFile(key string) string {
	// 使用SHA256哈希作为文件名，避免特殊字符问题
	hash := sha256.Sum256([]byte(key))
	filename := fmt.Sprintf("%x.json", hash)
	return filepath.Join(vc.cacheDir, filename)
}

// CacheStats 缓存统计信息
type CacheStats struct {
	CacheDir     string        `json:"cacheDir"`
	TTL          time.Duration `json:"ttl"`
	Enabled      bool          `json:"enabled"`
	TotalFiles   int           `json:"totalFiles"`
	ValidFiles   int           `json:"validFiles"`
	ExpiredFiles int           `json:"expiredFiles"`
	TotalSize    int64         `json:"totalSize"`
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(scanType, target string, options *ScanOptions) string {
	// 构建缓存键，包含扫描类型、目标和关键选项
	key := fmt.Sprintf("%s:%s", scanType, target)
	
	if options != nil {
		if len(options.SeverityFilter) > 0 {
			key += fmt.Sprintf(":severity=%v", options.SeverityFilter)
		}
		if !options.IncludeConfig {
			key += ":no-config"
		}
	}
	
	return key
}
