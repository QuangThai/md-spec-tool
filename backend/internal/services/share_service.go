package services

import (
	"context"
	"errors"

	"github.com/yourorg/md-spec-tool/internal/models"
	"github.com/yourorg/md-spec-tool/internal/repositories"
)

type ShareService struct {
	shareRepo *repositories.ShareRepository
	specRepo  *repositories.SpecRepository
	userRepo  *repositories.UserRepository
}

func NewShareService(shareRepo *repositories.ShareRepository, specRepo *repositories.SpecRepository, userRepo *repositories.UserRepository) *ShareService {
	return &ShareService{
		shareRepo: shareRepo,
		specRepo:  specRepo,
		userRepo:  userRepo,
	}
}

// ShareSpec shares a spec with another user
func (s *ShareService) ShareSpec(ctx context.Context, specID, ownerID, userID, permissionLevel string) (*models.Share, error) {
	// Verify owner owns the spec
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return nil, err
	}

	if spec.UserID != ownerID {
		return nil, errors.New("only spec owner can share")
	}

	// Verify the user exists
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.New("user not found")
	}

	// Can't share with self
	if userID == ownerID {
		return nil, errors.New("cannot share with self")
	}

	share := &models.Share{
		SpecID:           specID,
		SharedWithUserID: userID,
		PermissionLevel:  permissionLevel,
	}

	err = s.shareRepo.Create(ctx, share)
	if err != nil {
		return nil, err
	}

	return share, nil
}

// UnshareSpec removes a share
func (s *ShareService) UnshareSpec(ctx context.Context, specID, ownerID, userID string) error {
	// Verify owner owns the spec
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return err
	}

	if spec.UserID != ownerID {
		return errors.New("only spec owner can unshare")
	}

	return s.shareRepo.Delete(ctx, specID, userID)
}

// GetSpecShares returns all shares for a spec
func (s *ShareService) GetSpecShares(ctx context.Context, specID, ownerID string) ([]*models.Share, error) {
	// Verify owner owns the spec
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return nil, err
	}

	if spec.UserID != ownerID {
		return nil, errors.New("only spec owner can view shares")
	}

	return s.shareRepo.GetBySpecID(ctx, specID)
}

// CanAccessSpec checks if a user can access a spec (owner or shared with)
func (s *ShareService) CanAccessSpec(ctx context.Context, specID, userID string) (bool, string, error) {
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return false, "", err
	}

	// Owner has full access
	if spec.UserID == userID {
		return true, "owner", nil
	}

	// Check if shared with this user
	permission, err := s.shareRepo.GetPermission(ctx, specID, userID)
	if err != nil {
		return false, "", err
	}

	if permission != "" {
		return true, permission, nil
	}

	return false, "", nil
}

// UpdateSharePermission updates share permission level
func (s *ShareService) UpdateSharePermission(ctx context.Context, specID, ownerID, userID, permissionLevel string) error {
	// Verify owner owns the spec
	spec, err := s.specRepo.GetByID(ctx, specID)
	if err != nil {
		return err
	}

	if spec.UserID != ownerID {
		return errors.New("only spec owner can update permissions")
	}

	return s.shareRepo.UpdatePermission(ctx, specID, userID, permissionLevel)
}

// GetSharedSpecs returns all specs shared with a user
func (s *ShareService) GetSharedSpecs(ctx context.Context, userID string) ([]*models.Spec, error) {
	return s.shareRepo.GetSharedSpecsForUser(ctx, userID)
}
