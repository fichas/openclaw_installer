// Package version 提供版本比较功能
package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// versionRegex 匹配语义化版本格式
var versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$`)

// Version 表示语义化版本
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Metadata   string
	Original   string
}

// ParseVersion 解析版本字符串
func ParseVersion(v string) (*Version, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, fmt.Errorf("empty version string")
	}

	// 移除前缀 'v' 或 'V'
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")

	matches := versionRegex.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Metadata:   matches[5],
		Original:   v,
	}, nil
}

// String 返回版本字符串
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}
	if v.Metadata != "" {
		result += "+" + v.Metadata
	}
	return result
}

// Compare 比较两个版本
// 返回 -1 表示 v < other, 0 表示相等, 1 表示 v > other
func (v *Version) Compare(other *Version) int {
	// 比较主版本号
	if v.Major != other.Major {
		if v.Major > other.Major {
			return 1
		}
		return -1
	}

	// 比较次版本号
	if v.Minor != other.Minor {
		if v.Minor > other.Minor {
			return 1
		}
		return -1
	}

	// 比较修订版本号
	if v.Patch != other.Patch {
		if v.Patch > other.Patch {
			return 1
		}
		return -1
	}

	// 比较预发布版本
	return comparePrerelease(v.Prerelease, other.Prerelease)
}

// comparePrerelease 比较预发布版本
// 规则：没有预发布版本 > 有预发布版本
func comparePrerelease(a, b string) int {
	// 两者都没有预发布版本
	if a == "" && b == "" {
		return 0
	}

	// 没有预发布版本 > 有预发布版本
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}

	// 分割预发布标识符
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// 逐个比较
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aPart := aParts[i]
		bPart := bParts[i]

		// 尝试作为数字比较
		aNum, aErr := strconv.Atoi(aPart)
		bNum, bErr := strconv.Atoi(bPart)

		if aErr == nil && bErr == nil {
			// 都是数字
			if aNum != bNum {
				if aNum > bNum {
					return 1
				}
				return -1
			}
		} else {
			// 至少有一个不是数字，按字符串比较
			cmp := strings.Compare(aPart, bPart)
			if cmp != 0 {
				return cmp
			}
		}
	}

	// 较短的预发布版本 < 较长的
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}

	return 0
}

// CompareVersions 比较两个版本字符串
// 返回 -1 表示 v1 < v2, 0 表示相等, 1 表示 v1 > v2
func CompareVersions(v1, v2 string) int {
	ver1, err1 := ParseVersion(v1)
	ver2, err2 := ParseVersion(v2)

	// 如果解析失败，使用字符串比较
	if err1 != nil || err2 != nil {
		return strings.Compare(v1, v2)
	}

	return ver1.Compare(ver2)
}

// IsValidVersion 检查版本字符串是否有效
func IsValidVersion(v string) bool {
	_, err := ParseVersion(v)
	return err == nil
}

// IsPrerelease 检查是否为预发布版本
func IsPrerelease(v string) bool {
	ver, err := ParseVersion(v)
	if err != nil {
		return false
	}
	return ver.Prerelease != ""
}

// GetChannel 获取版本通道
func GetChannel(v string) string {
	ver, err := ParseVersion(v)
	if err != nil {
		return "unknown"
	}

	if ver.Prerelease == "" {
		return "stable"
	}

	pre := strings.ToLower(ver.Prerelease)
	if strings.Contains(pre, "alpha") {
		return "alpha"
	}
	if strings.Contains(pre, "beta") {
		return "beta"
	}
	if strings.Contains(pre, "rc") {
		return "rc"
	}

	return "dev"
}

// SatisfiesConstraint 检查版本是否满足约束条件
// 支持的约束格式:
//   - ">= 1.0.0"
//   - "^1.0.0" (兼容 1.x.x)
//   - "~1.0.0" (兼容 1.0.x)
//   - ">=1.0.0, <2.0.0"
func SatisfiesConstraint(version, constraint string) (bool, error) {
	ver, err := ParseVersion(version)
	if err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}

	constraint = strings.TrimSpace(constraint)

	// 处理 ^ 约束 (兼容版本)
	if strings.HasPrefix(constraint, "^") {
		baseVer, err := ParseVersion(constraint[1:])
		if err != nil {
			return false, err
		}
		// ^1.2.3 表示 >=1.2.3, <2.0.0
		lower := baseVer.Compare(ver) <= 0
		upper := ver.Major < baseVer.Major+1
		return lower && upper, nil
	}

	// 处理 ~ 约束 (近似版本)
	if strings.HasPrefix(constraint, "~") {
		baseVer, err := ParseVersion(constraint[1:])
		if err != nil {
			return false, err
		}
		// ~1.2.3 表示 >=1.2.3, <1.3.0
		lower := baseVer.Compare(ver) <= 0
		upper := ver.Major == baseVer.Major && ver.Minor < baseVer.Minor+1
		return lower && upper, nil
	}

	// 处理比较运算符
	if strings.HasPrefix(constraint, ">=") {
		baseVer, err := ParseVersion(constraint[2:])
		if err != nil {
			return false, err
		}
		return baseVer.Compare(ver) <= 0, nil
	}

	if strings.HasPrefix(constraint, ">") {
		baseVer, err := ParseVersion(constraint[1:])
		if err != nil {
			return false, err
		}
		return baseVer.Compare(ver) < 0, nil
	}

	if strings.HasPrefix(constraint, "<=") {
		baseVer, err := ParseVersion(constraint[2:])
		if err != nil {
			return false, err
		}
		return baseVer.Compare(ver) >= 0, nil
	}

	if strings.HasPrefix(constraint, "<") {
		baseVer, err := ParseVersion(constraint[1:])
		if err != nil {
			return false, err
		}
		return baseVer.Compare(ver) > 0, nil
	}

	if strings.HasPrefix(constraint, "=") {
		baseVer, err := ParseVersion(constraint[1:])
		if err != nil {
			return false, err
		}
		return baseVer.Compare(ver) == 0, nil
	}

	// 默认精确匹配
	baseVer, err := ParseVersion(constraint)
	if err != nil {
		return false, err
	}
	return baseVer.Compare(ver) == 0, nil
}
