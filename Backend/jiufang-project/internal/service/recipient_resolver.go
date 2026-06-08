package service

import (
	"context"
	"encoding/json"
	"strconv"

	"go.uber.org/zap"

	"jiufang/internal/repository"
)

// RecipientResolver resolves user snowflake IDs to email addresses.
type RecipientResolver struct {
	userRepo repository.UserRepositoryInterface
	logger   *zap.Logger
}

// NewRecipientResolver creates a new RecipientResolver instance.
func NewRecipientResolver(userRepo repository.UserRepositoryInterface, logger *zap.Logger) *RecipientResolver {
	return &RecipientResolver{userRepo: userRepo, logger: logger}
}

// ResolveEmails takes a JSON string of snowflake ID strings and returns the corresponding email addresses.
func (r *RecipientResolver) ResolveEmails(ctx context.Context, recipientsJSON string) ([]string, error) {
	var ids []string
	if err := json.Unmarshal([]byte(recipientsJSON), &ids); err != nil {
		return nil, err
	}

	var emails []string
	for _, idStr := range ids {
		sfID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.logger.Warn("invalid snowflake ID in recipients", zap.String("id", idStr))
			continue
		}
		u, err := r.userRepo.GetBySnowflakeID(ctx, sfID)
		if err != nil || u == nil {
			r.logger.Warn("recipient user not found", zap.Int64("snowflake_id", sfID))
			continue
		}
		if u.Email != "" {
			emails = append(emails, u.Email)
		}
	}
	return emails, nil
}
