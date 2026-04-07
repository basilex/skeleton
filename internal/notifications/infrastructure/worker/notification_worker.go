package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/basilex/skeleton/internal/notifications/domain"
	"github.com/basilex/skeleton/internal/notifications/infrastructure/sender"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type NotificationWorker struct {
	notificationRepo domain.NotificationRepository
	preferencesRepo  domain.PreferencesRepository
	compositeSender  *sender.CompositeSender
	eventBus         eventbus.Bus
	pollInterval     time.Duration
	batchSize        int
	stalledTimeout   time.Duration
}

type WorkerConfig struct {
	PollInterval   time.Duration
	BatchSize      int
	StalledTimeout time.Duration
}

func NewNotificationWorker(
	notificationRepo domain.NotificationRepository,
	preferencesRepo domain.PreferencesRepository,
	compositeSender *sender.CompositeSender,
	eventBus eventbus.Bus,
	config WorkerConfig,
) *NotificationWorker {
	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	batchSize := config.BatchSize
	if batchSize == 0 {
		batchSize = 100
	}

	stalledTimeout := config.StalledTimeout
	if stalledTimeout == 0 {
		stalledTimeout = 5 * time.Minute
	}

	return &NotificationWorker{
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		compositeSender:  compositeSender,
		eventBus:         eventBus,
		pollInterval:     pollInterval,
		batchSize:        batchSize,
		stalledTimeout:   stalledTimeout,
	}
}

func (w *NotificationWorker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	log.Println("Notification worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Notification worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processPending(ctx); err != nil {
				log.Printf("Error processing pending notifications: %v", err)
			}
			if err := w.processScheduled(ctx); err != nil {
				log.Printf("Error processing scheduled notifications: %v", err)
			}
			if err := w.processStalled(ctx); err != nil {
				log.Printf("Error processing stalled notifications: %v", err)
			}
		}
	}
}

func (w *NotificationWorker) processPending(ctx context.Context) error {
	notifications, err := w.notificationRepo.GetByStatus(ctx, domain.StatusPending, w.batchSize)
	if err != nil {
		return fmt.Errorf("get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := w.sendNotification(ctx, notification); err != nil {
			log.Printf("Failed to send notification %s: %v", notification.ID(), err)
		}
	}

	return nil
}

func (w *NotificationWorker) processScheduled(ctx context.Context) error {
	notifications, err := w.notificationRepo.GetScheduled(ctx, time.Now(), w.batchSize)
	if err != nil {
		return fmt.Errorf("get scheduled notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := w.sendNotification(ctx, notification); err != nil {
			log.Printf("Failed to send scheduled notification %s: %v", notification.ID(), err)
		}
	}

	return nil
}

func (w *NotificationWorker) processStalled(ctx context.Context) error {
	notifications, err := w.notificationRepo.GetStalled(ctx, w.stalledTimeout, w.batchSize)
	if err != nil {
		return fmt.Errorf("get stalled notifications: %w", err)
	}

	for _, notification := range notifications {
		log.Printf("Retrying stalled notification %s (attempt %d)", notification.ID(), notification.Attempts())
		if err := w.sendNotification(ctx, notification); err != nil {
			log.Printf("Failed to retry stalled notification %s: %v", notification.ID(), err)
		}
	}

	return nil
}

func (w *NotificationWorker) sendNotification(ctx context.Context, notification *domain.Notification) error {
	if notification.Recipient().UserID != nil {
		preferences, err := w.preferencesRepo.GetByUserID(ctx, string(*notification.Recipient().UserID))
		if err == nil {
			if !preferences.IsChannelEnabled(notification.Channel()) {
				log.Printf("Channel %s disabled for user %s, skipping notification %s",
					notification.Channel(), *notification.Recipient().UserID, notification.ID())
				_ = notification.MarkFailed("channel disabled by user preferences")
				_ = w.notificationRepo.Update(ctx, notification)
				return nil
			}

			if !preferences.CanSendNow(notification.Channel(), time.Now()) {
				log.Printf("Cannot send notification %s due to quiet hours, rescheduling", notification.ID())
				_ = notification.ScheduleRetry(1 * time.Hour)
				_ = w.notificationRepo.Update(ctx, notification)
				return nil
			}
		}
	}

	if err := notification.StartSending(); err != nil {
		return fmt.Errorf("start sending: %w", err)
	}
	notification.IncrementAttempts()

	if err := w.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update notification status: %w", err)
	}

	if err := w.compositeSender.Send(ctx, notification); err != nil {
		return w.handleSendError(ctx, notification, err)
	}

	if err := notification.MarkSent(); err != nil {
		return fmt.Errorf("mark as sent: %w", err)
	}

	if err := w.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update notification: %w", err)
	}

	event := domain.NewNotificationSent(notification.ID(), notification.Channel())
	w.eventBus.Publish(ctx, event)

	log.Printf("Successfully sent notification %s via %s", notification.ID(), notification.Channel())
	return nil
}

func (w *NotificationWorker) handleSendError(ctx context.Context, notification *domain.Notification, sendErr error) error {
	if notification.CanRetry() {
		delay := notification.NextRetryDelay()
		if err := notification.ScheduleRetry(delay); err != nil {
			return fmt.Errorf("schedule retry: %w", err)
		}

		if err := w.notificationRepo.Update(ctx, notification); err != nil {
			return fmt.Errorf("update notification for retry: %w", err)
		}

		event := domain.NewNotificationFailed(
			notification.ID(),
			notification.Channel(),
			sendErr.Error(),
			notification.Attempts(),
			true,
			notification.ScheduledAt(),
		)
		w.eventBus.Publish(ctx, event)

		log.Printf("Scheduled retry for notification %s in %v (attempt %d/%d)",
			notification.ID(), delay, notification.Attempts(), notification.MaxAttempts())
		return nil
	}

	_ = notification.MarkFailed(sendErr.Error())

	if err := w.notificationRepo.Update(ctx, notification); err != nil {
		return fmt.Errorf("update failed notification: %w", err)
	}

	event := domain.NewNotificationFailed(
		notification.ID(),
		notification.Channel(),
		sendErr.Error(),
		notification.Attempts(),
		false,
		nil,
	)
	w.eventBus.Publish(ctx, event)

	log.Printf("Notification %s failed after %d attempts: %v", notification.ID(), notification.MaxAttempts(), sendErr)
	return fmt.Errorf("notification failed: %w", sendErr)
}
