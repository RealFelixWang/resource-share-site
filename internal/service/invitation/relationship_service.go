/*
Package invitation provides invitation relationship tracking services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package invitation

import (
	"fmt"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// InvitationTreeNode 邀请树节点
type InvitationTreeNode struct {
	User        *model.User           `json:"user"`
	Invitations []*model.Invitation   `json:"invitations"`
	Children    []*InvitationTreeNode `json:"children,omitempty"`
}

// InvitationPath 邀请路径
type InvitationPath struct {
	Inviter      *model.User `json:"inviter"`
	Invitee      *model.User `json:"invitee"`
	InvitedAt    time.Time   `json:"invited_at"`
	PointsEarned int         `json:"points_earned"`
}

// RelationshipService 邀请关系服务
type RelationshipService struct {
	db *gorm.DB
}

// NewRelationshipService 创建新的邀请关系服务
func NewRelationshipService(db *gorm.DB) *RelationshipService {
	return &RelationshipService{
		db: db,
	}
}

// GetUserInvitedUsers 获取用户直接邀请的用户列表
// 参数：
//   - userID: 用户ID
//   - page: 页码
//   - pageSize: 每页数量
//
// 返回：
//   - 被邀请的用户列表
//   - 总数
//   - 错误信息
func (s *RelationshipService) GetUserInvitedUsers(userID uint, page, pageSize int) ([]*model.User, int64, error) {
	// 通过InvitedByID关联查询
	var users []*model.User
	query := s.db.Model(&model.User{}).Where("invited_by_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询被邀请用户总数失败: %w", err)
	}

	if err := query.Preload("InvitedBy").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("查询被邀请用户列表失败: %w", err)
	}

	return users, total, nil
}

// GetUserInvitationTree 获取用户的邀请树结构（最多3层）
// 参数：
//   - userID: 用户ID
//   - maxDepth: 最大深度
//
// 返回：
//   - 邀请树根节点
//   - 错误信息
func (s *RelationshipService) GetUserInvitationTree(userID uint, maxDepth int) (*InvitationTreeNode, error) {
	if maxDepth <= 0 || maxDepth > 5 {
		maxDepth = 3 // 默认最多3层
	}

	// 获取根用户信息
	var rootUser model.User
	if err := s.db.First(&rootUser, userID).Error; err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 构建树结构
	root := &InvitationTreeNode{
		User: &rootUser,
	}

	// 递归获取子节点
	if err := s.buildTree(root, 1, maxDepth); err != nil {
		return nil, fmt.Errorf("构建邀请树失败: %w", err)
	}

	return root, nil
}

// GetUserInvitationPath 获取用户的邀请路径（从根到用户）
// 参数：
//   - userID: 用户ID
//
// 返回：
//   - 邀请路径
//   - 错误信息
func (s *RelationshipService) GetUserInvitationPath(userID uint) ([]*InvitationPath, error) {
	var path []*InvitationPath

	// 递归向上查找邀请关系
	currentUserID := userID
	for {
		// 获取当前用户及其邀请者
		var user model.User
		if err := s.db.Preload("InvitedBy").First(&user, currentUserID).Error; err != nil {
			break // 已到达根用户
		}

		// 如果用户没有被邀请，则到达根用户
		if user.InvitedByID == nil {
			break
		}

		// 查找对应的邀请记录
		var invitation model.Invitation
		if err := s.db.Where("inviter_id = ? AND invitee_id = ? AND status = ?",
			*user.InvitedByID, user.ID, model.InvitationStatusCompleted).
			First(&invitation).Error; err != nil {
			break
		}

		// 添加到路径
		path = append([]*InvitationPath{{
			Inviter:      user.InvitedBy,
			Invitee:      &user,
			InvitedAt:    invitation.CreatedAt,
			PointsEarned: invitation.PointsAwarded,
		}}, path...)

		// 移动到邀请者
		currentUserID = *user.InvitedByID
	}

	return path, nil
}

// GetInvitationCountByLevel 获取用户各层级的邀请数量
// 参数：
//   - userID: 用户ID
//   - maxLevel: 最大层级
//
// 返回：
//   - 各层级邀请数量
//   - 错误信息
func (s *RelationshipService) GetInvitationCountByLevel(userID uint, maxLevel int) (map[int]int64, error) {
	if maxLevel <= 0 || maxLevel > 5 {
		maxLevel = 3
	}

	counts := make(map[int]int64)
	for level := 1; level <= maxLevel; level++ {
		count, err := s.countUsersAtLevel(userID, level)
		if err != nil {
			return nil, fmt.Errorf("查询第%d层邀请数量失败: %w", level, err)
		}
		counts[level] = count
	}

	return counts, nil
}

// GetTopInviters 获取邀请排行榜
// 参数：
//   - limit: 限制数量
//   - timeRange: 时间范围（"all", "month", "week", "day"）
//
// 返回：
//   - 排行榜列表
//   - 错误信息
func (s *RelationshipService) GetTopInviters(limit int, timeRange string) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	// 构建查询
	query := s.db.Table("users").
		Select(`users.id, users.username, users.email,
				COUNT(invitations.id) as invite_count,
				SUM(invitations.points_awarded) as total_points`).
		Joins("LEFT JOIN invitations ON users.id = invitations.inviter_id AND invitations.status = ?", model.InvitationStatusCompleted).
		Group("users.id")

	// 添加时间范围筛选
	switch timeRange {
	case "month":
		query = query.Where("invitations.created_at >= ?", time.Now().AddDate(0, -1, 0))
	case "week":
		query = query.Where("invitations.created_at >= ?", time.Now().AddDate(0, 0, -7))
	case "day":
		query = query.Where("invitations.created_at >= ?", time.Now().AddDate(0, 0, -1))
	default: // "all"
		// 不添加时间限制
	}

	// 排序和限制
	var results []map[string]interface{}
	if err := query.Order("invite_count DESC, total_points DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("查询邀请排行榜失败: %w", err)
	}

	return results, nil
}

// GetNetworkStats 获取网络统计信息
// 参数：
//   - userID: 用户ID
//
// 返回：
//   - 网络统计信息
//   - 错误信息
func (s *RelationshipService) GetNetworkStats(userID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 直接邀请用户数
	var directInvites int64
	if err := s.db.Model(&model.User{}).Where("invited_by_id = ?", userID).Count(&directInvites).Error; err != nil {
		return nil, fmt.Errorf("查询直接邀请数失败: %w", err)
	}
	stats["direct_invites"] = directInvites

	// 第二层邀请用户数
	var secondLevelInvites int64
	if err := s.db.Raw(`
		SELECT COUNT(*)
		FROM users u1
		JOIN users u2 ON u1.id = u2.invited_by_id
		WHERE u1.invited_by_id = ?
	`, userID).Scan(&secondLevelInvites).Error; err != nil {
		return nil, fmt.Errorf("查询第二层邀请数失败: %w", err)
	}
	stats["second_level_invites"] = secondLevelInvites

	// 第三层邀请用户数
	var thirdLevelInvites int64
	if err := s.db.Raw(`
		SELECT COUNT(*)
		FROM users u1
		JOIN users u2 ON u1.id = u2.invited_by_id
		JOIN users u3 ON u2.id = u3.invited_by_id
		WHERE u1.invited_by_id = ?
	`, userID).Scan(&thirdLevelInvites).Error; err != nil {
		return nil, fmt.Errorf("查询第三层邀请数失败: %w", err)
	}
	stats["third_level_invites"] = thirdLevelInvites

	// 总邀请用户数
	totalInvites := directInvites + secondLevelInvites + thirdLevelInvites
	stats["total_network_size"] = totalInvites

	// 活跃用户数（近30天内有活动的用户）
	var activeInvites int64
	if err := s.db.Raw(`
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE u.invited_by_id IN (
			SELECT id FROM users WHERE invited_by_id = ?
		)
		AND u.updated_at >= ?
	`, userID, time.Now().AddDate(0, 0, -30)).Scan(&activeInvites).Error; err != nil {
		return nil, fmt.Errorf("查询活跃邀请数失败: %w", err)
	}
	stats["active_invites"] = activeInvites

	// 邀请网络深度
	depth, err := s.getNetworkDepth(userID)
	if err != nil {
		return nil, fmt.Errorf("查询网络深度失败: %w", err)
	}
	stats["network_depth"] = depth

	return stats, nil
}

// buildTree 递归构建树结构
func (s *RelationshipService) buildTree(node *InvitationTreeNode, currentDepth, maxDepth int) error {
	if currentDepth >= maxDepth {
		return nil
	}

	// 获取子节点
	var children []*model.User
	if err := s.db.Where("invited_by_id = ?", node.User.ID).Find(&children).Error; err != nil {
		return err
	}

	// 为每个子节点创建节点
	for _, child := range children {
		childNode := &InvitationTreeNode{
			User: child,
		}

		// 递归添加子节点
		if err := s.buildTree(childNode, currentDepth+1, maxDepth); err != nil {
			return err
		}

		node.Children = append(node.Children, childNode)
	}

	return nil
}

// countUsersAtLevel 计算指定层级的用户数量
func (s *RelationshipService) countUsersAtLevel(userID uint, level int) (int64, error) {
	if level == 1 {
		// 第一层：直接邀请的用户
		var count int64
		err := s.db.Model(&model.User{}).Where("invited_by_id = ?", userID).Count(&count).Error
		return count, err
	}

	// 多层级查询使用原生SQL
	query := fmt.Sprintf(`
		WITH RECURSIVE invite_tree AS (
			SELECT id, invited_by_id, 1 as level
			FROM users
			WHERE invited_by_id = ?
			UNION ALL
			SELECT u.id, u.invited_by_id, it.level + 1
			FROM users u
			INNER JOIN invite_tree it ON u.invited_by_id = it.id
			WHERE it.level < ?
		)
		SELECT COUNT(*) FROM invite_tree WHERE level = ?
	`, userID, level, level)

	var count int64
	err := s.db.Raw(query).Scan(&count).Error
	return count, err
}

// getNetworkDepth 获取网络深度
func (s *RelationshipService) getNetworkDepth(userID uint) (int, error) {
	query := `
		WITH RECURSIVE invite_tree AS (
			SELECT id, invited_by_id, 1 as level
			FROM users
			WHERE invited_by_id = ?
			UNION ALL
			SELECT u.id, u.invited_by_id, it.level + 1
			FROM users u
			INNER JOIN invite_tree it ON u.invited_by_id = it.id
		)
		SELECT MAX(level) FROM invite_tree
	`

	var depth int
	err := s.db.Raw(query, userID).Scan(&depth).Error
	if err != nil {
		return 0, err
	}
	return depth, nil
}
