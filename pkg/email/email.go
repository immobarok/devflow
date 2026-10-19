package email

import (
	"context"
	"log/slog"

	"github.com/mobarok/devflow/pkg/worker"
)

type Service interface {
	SendInvitation(toEmail, inviterName, orgName, inviteLink string)
	SendPasswordReset(toEmail, otp string)
	SendTaskAssigned(toEmail, taskKey, taskTitle, assignerName string)
}

type emailService struct {
	pool *worker.Pool
}

func NewService(pool *worker.Pool) Service {
	return &emailService{pool: pool}
}

func (s *emailService) SendInvitation(toEmail, inviterName, orgName, inviteLink string) {
	job := func(ctx context.Context) error {
		slog.Info("📧 [EMAIL OUTBOX] Workspace Invitation",
			"to", toEmail,
			"inviter", inviterName,
			"organization", orgName,
			"link", inviteLink,
		)
		return nil
	}

	if s.pool != nil {
		s.pool.Submit(job)
	} else {
		_ = job(context.Background())
	}
}

func (s *emailService) SendPasswordReset(toEmail, otp string) {
	job := func(ctx context.Context) error {
		slog.Info("📧 [EMAIL OUTBOX] Password Reset OTP",
			"to", toEmail,
			"otp", otp,
		)
		return nil
	}

	if s.pool != nil {
		s.pool.Submit(job)
	} else {
		_ = job(context.Background())
	}
}

func (s *emailService) SendTaskAssigned(toEmail, taskKey, taskTitle, assignerName string) {
	job := func(ctx context.Context) error {
		slog.Info("📧 [EMAIL OUTBOX] Task Assigned",
			"to", toEmail,
			"task_key", taskKey,
			"title", taskTitle,
			"assigned_by", assignerName,
		)
		return nil
	}

	if s.pool != nil {
		s.pool.Submit(job)
	} else {
		_ = job(context.Background())
	}
}
