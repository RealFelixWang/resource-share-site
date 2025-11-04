/*
Package config provides configuration management for the resource share site.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package config

import "errors"

var (
	// ErrUnsupportedDBType 不支持的数据库类型
	ErrUnsupportedDBType = errors.New("不支持的数据库类型")

	// ErrConfigNotFound 配置未找到
	ErrConfigNotFound = errors.New("配置未找到")

	// ErrConfigInvalid 配置无效
	ErrConfigInvalid = errors.New("配置无效")

	// ErrRedisConfigNil Redis配置为空
	ErrRedisConfigNil = errors.New("Redis配置为空")

	// ErrRedisConnectionFailed Redis连接失败
	ErrRedisConnectionFailed = errors.New("Redis连接失败")

	// ErrRedisOperationFailed Redis操作失败
	ErrRedisOperationFailed = errors.New("Redis操作失败")
)
