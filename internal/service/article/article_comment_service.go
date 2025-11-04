/*
Package article provides article comment services.

Author: Felix Wang
Email: felixwang.biz@gmail.com
*/

package article

import (
	"errors"
	"time"

	"resource-share-site/internal/model"

	"gorm.io/gorm"
)

// ArticleCommentService 文章评论服务
type ArticleCommentService struct {
	db *gorm.DB
}

// NewArticleCommentService 创建文章评论服务实例
func NewArticleCommentService(db *gorm.DB) *ArticleCommentService {
	return &ArticleCommentService{
		db: db,
	}
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	ArticleID uint   `json:"article_id" binding:"required"`
	Content   string `json:"content" binding:"required,min=1,max=1000"`
	ParentID  *uint  `json:"parent_id"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// CommentListItem 评论列表项
type CommentListItem struct {
	ID          uint      `json:"id"`
	Content     string    `json:"content"`
	LikeCount   uint      `json:"like_count"`
	CreatedAt   time.Time `json:"created_at"`

	// 用户信息
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`

	// 父评论信息
	ParentID *uint  `json:"parent_id"`
}

// CreateComment 创建评论
func (s *ArticleCommentService) CreateComment(userID uint, req *CreateCommentRequest) (*model.ArticleComment, error) {
	// 检查文章是否存在
	var article model.Article
	if err := s.db.First(&article, req.ArticleID).Error; err != nil {
		return nil, err
	}

	// 如果有父评论，检查父评论是否存在且属于同一文章
	if req.ParentID != nil {
		var parentComment model.ArticleComment
		if err := s.db.First(&parentComment, *req.ParentID).Error; err != nil {
			return nil, err
		}
		if parentComment.ArticleID != req.ArticleID {
			return nil, errors.New("父评论不属于同一文章")
		}
	}

	// 创建评论
	comment := &model.ArticleComment{
		ArticleID: req.ArticleID,
		UserID:    userID,
		Content:   req.Content,
		ParentID:  req.ParentID,
		Status:    model.CommentStatusPending, // 评论需要审核
	}

	if err := s.db.Create(comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentByID 根据ID获取评论
func (s *ArticleCommentService) GetCommentByID(id uint) (*model.ArticleComment, error) {
	var comment model.ArticleComment
	if err := s.db.Preload("User").Preload("Parent").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// UpdateComment 更新评论
func (s *ArticleCommentService) UpdateComment(id uint, userID uint, req *UpdateCommentRequest) (*model.ArticleComment, error) {
	comment, err := s.GetCommentByID(id)
	if err != nil {
		return nil, err
	}

	// 检查是否是评论作者
	if comment.UserID != userID {
		return nil, errors.New("只能修改自己的评论")
	}

	// 只有待审核的评论才能修改
	if comment.Status != model.CommentStatusPending {
		return nil, errors.New("已审核的评论不能修改")
	}

	comment.Content = req.Content

	if err := s.db.Save(comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment 删除评论（软删除）
func (s *ArticleCommentService) DeleteComment(id uint, userID uint) error {
	comment, err := s.GetCommentByID(id)
	if err != nil {
		return err
	}

	// 检查是否是评论作者或管理员
	// TODO: 添加管理员检查逻辑

	if comment.UserID != userID {
		return errors.New("只能删除自己的评论")
	}

	if err := s.db.Delete(comment).Error; err != nil {
		return err
	}

	return nil
}

// ApproveComment 审核通过评论
func (s *ArticleCommentService) ApproveComment(id uint, reviewerID uint, reviewNotes string) error {
	comment, err := s.GetCommentByID(id)
	if err != nil {
		return err
	}

	comment.Status = model.CommentStatusApproved
	comment.ReviewedByID = &reviewerID
	now := time.Now()
	comment.ReviewedAt = &now
	comment.ReviewNotes = reviewNotes

	if err := s.db.Save(comment).Error; err != nil {
		return err
	}

	// 增加文章的评论数
	return s.db.Model(&model.Article{}).
		Where("id = ?", comment.ArticleID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

// RejectComment 审核拒绝评论
func (s *ArticleCommentService) RejectComment(id uint, reviewerID uint, reviewNotes string) error {
	comment, err := s.GetCommentByID(id)
	if err != nil {
		return err
	}

	comment.Status = model.CommentStatusRejected
	comment.ReviewedByID = &reviewerID
	now := time.Now()
	comment.ReviewedAt = &now
	comment.ReviewNotes = reviewNotes

	if err := s.db.Save(comment).Error; err != nil {
		return err
	}

	return nil
}

// GetCommentsByArticleID 获取文章评论列表
func (s *ArticleCommentService) GetCommentsByArticleID(articleID uint, page, pageSize int) ([]CommentListItem, int64, error) {
	var comments []CommentListItem
	var total int64

	// 构建查询
	query := s.db.Table("article_comments").
		Select(`
			article_comments.*,
			users.username,
			users.avatar
		`).
		Joins("LEFT JOIN users ON article_comments.user_id = users.id").
		Where("article_comments.article_id = ?", articleID).
		Where("article_comments.status = ?", model.CommentStatusApproved)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("article_comments.created_at ASC").
		Scan(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetCommentsByUserID 获取用户评论列表
func (s *ArticleCommentService) GetCommentsByUserID(userID uint, page, pageSize int) ([]CommentListItem, int64, error) {
	var comments []CommentListItem
	var total int64

	query := s.db.Table("article_comments").
		Select(`
			article_comments.*,
			users.username,
			users.avatar
		`).
		Joins("LEFT JOIN users ON article_comments.user_id = users.id").
		Where("article_comments.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("article_comments.created_at DESC").
		Scan(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetPendingComments 获取待审核评论
func (s *ArticleCommentService) GetPendingComments(page, pageSize int) ([]CommentListItem, int64, error) {
	var comments []CommentListItem
	var total int64

	query := s.db.Table("article_comments").
		Select(`
			article_comments.*,
			users.username,
			users.avatar
		`).
		Joins("LEFT JOIN users ON article_comments.user_id = users.id").
		Where("article_comments.status = ?", model.CommentStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("article_comments.created_at DESC").
		Scan(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// IncrementLikeCount 增加点赞数
func (s *ArticleCommentService) IncrementLikeCount(id uint) error {
	return s.db.Model(&model.ArticleComment{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// DecrementLikeCount 减少点赞数
func (s *ArticleCommentService) DecrementLikeCount(id uint) error {
	return s.db.Model(&model.ArticleComment{}).
		Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
}
