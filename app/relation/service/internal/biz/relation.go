package biz

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

type User struct {
	Id              uint32 // 用户id
	Name            string // 用户名称
	FollowCount     uint32 // 关注总数
	FollowerCount   uint32 // 粉丝总数
	IsFollow        bool   // true-已关注，false-未关注
	Avatar          string // 用户头像
	BackgroundImage string // 用户个人页顶部大图
	Signature       string // 个人简介
	TotalFavorite   uint32 // 获赞数量
	WorkCount       uint32 // 作品数量
	FavoriteCount   uint32 // 点赞数量
}

type RelationRepo interface {
	GetFollowList(context.Context, uint32) ([]*User, error)
	GetFollowerList(context.Context, uint32) ([]*User, error)
	Follow(context.Context, uint32) error
	UnFollow(context.Context, uint32) error
	IsFollow(ctx context.Context, userId uint32, toUserId []uint32) ([]bool, error)
}

type RelationUseCase struct {
	repo RelationRepo
	log  *log.Helper
}

func NewRelationUseCase(repo RelationRepo, logger log.Logger) *RelationUseCase {
	return &RelationUseCase{repo: repo, log: log.NewHelper(logger)}
}

// GetFollowList 获取关注列表
func (uc *RelationUseCase) GetFollowList(ctx context.Context, userId uint32) ([]*User, error) {
	return uc.repo.GetFollowList(ctx, userId)
}

// GetFollowerList 获取粉丝列表
func (uc *RelationUseCase) GetFollowerList(ctx context.Context, userId uint32) ([]*User, error) {
	return uc.repo.GetFollowerList(ctx, userId)
}

// Action 关注和取消关注
func (uc *RelationUseCase) Action(ctx context.Context, toUserId uint32, actionType uint32) error {
	var followType uint32 = 1
	var unfollowType uint32 = 2
	switch actionType {
	// 1为关注
	case followType:
		err := uc.repo.Follow(ctx, toUserId)
		if err != nil {
			return fmt.Errorf("failed to follow: %w", err)
		}
	// 2为取消关注
	case unfollowType:
		err := uc.repo.UnFollow(ctx, toUserId)
		if err != nil {
			return fmt.Errorf("failed to unfollow: %w", err)
		}
	}
	return nil
}

func (uc *RelationUseCase) IsFollow(ctx context.Context, userId uint32, toUserId []uint32) ([]bool, error) {
	return uc.repo.IsFollow(ctx, userId, toUserId)
}
