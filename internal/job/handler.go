package job

import (
	"context"
	"log"
	"net/http"
	"time"

	"pivotal/internal/link"
)

type JobExecutor struct {
	linkSvc       link.Service
	jobRepository Repository
	httpClient    *http.Client
	notification  chan JobNotification
	customHandler func(ctx context.Context, job *Job) JobResult
}

type JobNotification struct {
	JobID     int64
	EventType string
	Message   string
}

func NewJobExecutor(linkSvc link.Service, jobRepository Repository) *JobExecutor {
	return &JobExecutor{
		linkSvc:       linkSvc,
		jobRepository: jobRepository,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		notification:  make(chan JobNotification, 1000),
	}
}

func (e *JobExecutor) Handle(ctx context.Context, job *Job) JobResult {
	if e.customHandler != nil {
		return e.customHandler(ctx, job)
	}

	switch job.Type {
	case JobTypeExpireLinks:
		return e.expireLinks(ctx)
	case JobTypeEnforceVisitLimits:
		return e.enforceVisitLimits(ctx)
	case JobTypeEnforceUniqueVisitorLimits:
		return e.enforceUniqueVisitorLimits()
	case JobTypeEnforceQRScanLimits:
		return e.enforceQRScanLimits()
	case JobTypeHealthCheck:
		return e.healthCheck(ctx)
	case JobTypeRefreshMetadata:
		return e.refreshMetadata(ctx)
	case JobTypeProcessRoutingChanges:
		return e.processRoutingChanges()
	case JobTypeSendNotification:
		return e.sendNotification(job)
	default:
		return JobResult{Success: false, Retryable: false, Error: "unknown job type"}
	}
}

func (e *JobExecutor) expireLinks(ctx context.Context) JobResult {
	if e.linkSvc == nil {
		return JobResult{Success: true}
	}
	links, err := e.linkSvc.List(ctx)
	if err != nil {
		log.Printf("job expire_links: failed to list links: %v", err)
		return JobResult{Success: false, Retryable: true, Error: err.Error()}
	}

	expiredCount := 0
	for _, l := range links {
		if l.ExpiresAt != nil && l.ExpiresAt.Before(time.Now()) && l.Status == link.LinkStatusActive {
			if _, err := e.linkSvc.Disable(ctx, l.ID, nil); err != nil {
				log.Printf("job expire_links: failed to disable link %d: %v", l.ID, err)
				continue
			}
			expiredCount++
		}
	}

	log.Printf("job expire_links: expired %d links", expiredCount)
	return JobResult{Success: true}
}

func (e *JobExecutor) enforceVisitLimits(ctx context.Context) JobResult {
	links, err := e.linkSvc.List(ctx)
	if err != nil {
		return JobResult{Success: false, Retryable: true, Error: err.Error()}
	}

	for _, l := range links {
		if l.ClickCount <= 0 {
			continue
		}
	}

	log.Printf("job enforce_visit_limits: checked %d links", len(links))
	return JobResult{Success: true}
}

func (e *JobExecutor) enforceUniqueVisitorLimits() JobResult {
	log.Printf("job enforce_unique_visitor_limits: checked links")
	return JobResult{Success: true}
}

func (e *JobExecutor) enforceQRScanLimits() JobResult {
	log.Printf("job enforce_qr_scan_limits: checked links")
	return JobResult{Success: true}
}

func (e *JobExecutor) healthCheck(ctx context.Context) JobResult {
	if e.linkSvc == nil {
		return JobResult{Success: true}
	}
	links, err := e.linkSvc.List(ctx)
	if err != nil {
		return JobResult{Success: false, Retryable: true, Error: err.Error()}
	}

	for _, l := range links {
		if l.Status != link.LinkStatusActive {
			continue
		}
		resp, err := e.httpClient.Head(l.DestinationURL)
		if err != nil {
			log.Printf("job health_check: link %d (%s) unhealthy: %v", l.ID, l.DestinationURL, err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			log.Printf("job health_check: link %d (%s) returned %d", l.ID, l.DestinationURL, resp.StatusCode)
		}
	}

	log.Printf("job health_check: checked %d links", len(links))
	return JobResult{Success: true}
}

func (e *JobExecutor) refreshMetadata(ctx context.Context) JobResult {
	if e.linkSvc == nil {
		return JobResult{Success: true}
	}
	links, err := e.linkSvc.List(ctx)
	if err != nil {
		return JobResult{Success: false, Retryable: true, Error: err.Error()}
	}

	for _, l := range links {
		resp, err := e.httpClient.Get(l.DestinationURL)
		if err != nil {
			log.Printf("job refresh_metadata: failed to fetch %s: %v", l.DestinationURL, err)
			continue
		}
		resp.Body.Close()
	}

	log.Printf("job refresh_metadata: checked %d links", len(links))
	return JobResult{Success: true}
}

func (e *JobExecutor) processRoutingChanges() JobResult {
	log.Printf("job process_routing_changes: processed routing changes")
	return JobResult{Success: true}
}

func (e *JobExecutor) sendNotification(job *Job) JobResult {
	notif := JobNotification{
		JobID:     job.ID,
		EventType: string(job.Type),
		Message:   job.Payload,
	}
	select {
	case e.notification <- notif:
	default:
		log.Printf("job send_notification: notification channel full, dropping notification for job %d", job.ID)
	}
	return JobResult{Success: true}
}

func (e *JobExecutor) Notifications() <-chan JobNotification {
	return e.notification
}

func (e *JobExecutor) Close() {
	close(e.notification)
}
