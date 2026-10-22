package models

import (
	"encoding/json"
	"errors"
	pb "github.com/alexQi/kerthus-core/rpc/saas/v1"
	microerrors "go-micro.dev/v6/errors"
	"time"
)

var ErrSchedulerUnsupported = errors.New("model scheduler v1 unavailable; do not fall back to a paid legacy submission")

var ErrDailyQuota = errors.New("model daily quota exhausted")

type AdmissionError struct {
	Code       string
	RetryAfter time.Duration
	NotSent    bool
}

func (e *AdmissionError) Error() string { return e.Code }
func (e *AdmissionError) Is(target error) bool {
	return target == ErrDailyQuota && e.Code == "model_daily_quota_exhausted"
}
func admissionResult(reply *pb.ModelCallReply, err error) (*pb.ModelCallReply, error) {
	if err == nil {
		return reply, nil
	}
	parsed := microerrors.Parse(err.Error())
	if parsed.Id == "kerthus.saas" && parsed.Code == 429 && parsed.Detail == "当前应用模型的每日调用额度已用完" {
		return reply, &AdmissionError{Code: "model_daily_quota_exhausted", RetryAfter: time.Minute, NotSent: true}
	}
	if parsed.Id == "kerthus.model.admission" && parsed.Code == 429 {
		var v struct {
			Code       string `json:"code"`
			RetryAfter int64  `json:"retry_after_seconds"`
			NotSent    bool   `json:"not_sent"`
		}
		if json.Unmarshal([]byte(parsed.Detail), &v) == nil && v.NotSent && (v.Code == "model_capacity_exhausted" || v.Code == "model_queue_full" || v.Code == "model_daily_quota_exhausted") {
			return reply, &AdmissionError{Code: v.Code, RetryAfter: time.Duration(v.RetryAfter) * time.Second, NotSent: true}
		}
	}
	return reply, err
}
