package collector

import (
	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/toolkits/pkg/logger"
)

type TargetRegistrar struct {
	ctx       *ctx.Context
	groupId   int64
	engineName string
}

func NewTargetRegistrar(ctx *ctx.Context, groupId int64, engineName string) *TargetRegistrar {
	return &TargetRegistrar{
		ctx:        ctx,
		groupId:    groupId,
		engineName: engineName,
	}
}

func (r *TargetRegistrar) RegisterAlertSource(event *AlertEvent) error {
	target := &models.Target{
		GroupId:    r.groupId,
		Ident:      event.Ident,
		EngineName: r.engineName,
		HostIp:     getHostIpFromLabels(event.Labels),
		Note:       "Alert source: " + event.Source,
		TagsJSON:   extractTags(event.Labels),
	}

	existingTarget, err := models.TargetGetByIdent(r.ctx, event.Ident)
	if err != nil {
		logger.Errorf("failed to get existing target for ident %s: %v", event.Ident, err)
		return err
	}

	if existingTarget != nil {
		existingTarget.UpdateAt = models.UpdateFields{}
		existingTarget.TagsJSON = target.TagsJSON
		existingTarget.Note = target.Note

		err = existingTarget.Update(r.ctx)
		if err != nil {
			logger.Errorf("failed to update target for ident %s: %v", event.Ident, err)
			return err
		}
		logger.Infof("updated existing target for ident %s", event.Ident)
	} else {
		err = target.Save(r.ctx)
		if err != nil {
			logger.Errorf("failed to create target for ident %s: %v", event.Ident, err)
			return err
		}
		logger.Infof("created new target for ident %s", event.Ident)
	}

	return nil
}

func (r *TargetRegistrar) BatchRegisterAlertSources(events []AlertEvent) (int, int) {
	successCount := 0
	errorCount := 0

	registeredIdents := make(map[string]bool)

	for _, event := range events {
		if registeredIdents[event.Ident] {
			continue
		}

		err := r.RegisterAlertSource(&event)
		if err != nil {
			errorCount++
		} else {
			successCount++
			registeredIdents[event.Ident] = true
		}
	}

	return successCount, errorCount
}

func getHostIpFromLabels(labels map[string]string) string {
	if ip, ok := labels["host_ip"]; ok {
		return ip
	}
	if ip, ok := labels["instance"]; ok {
		return ip
	}
	if ip, ok := labels["ip"]; ok {
		return ip
	}
	return ""
}

func extractTags(labels map[string]string) []string {
	tags := make([]string, 0, len(labels))

	excludeKeys := map[string]bool{
		"host_ip":  true,
		"instance": true,
		"ip":       true,
		"ident":    true,
		"alert_name": true,
		"severity": true,
	}

	for k, v := range labels {
		if !excludeKeys[k] && v != "" {
			tags = append(tags, k+"="+v)
		}
	}

	return tags
}
