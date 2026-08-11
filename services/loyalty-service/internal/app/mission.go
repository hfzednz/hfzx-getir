package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// TrackMissionInput increments mission progress.
type TrackMissionInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
	MissionID uuid.UUID
	Delta     int64
}

// TrackMission increments progress; auto-completes when target reached.
func (d *Deps) TrackMission(ctx context.Context, in TrackMissionInput) (domain.MissionProgress, error) {
	if in.Delta <= 0 {
		in.Delta = 1
	}
	mission, err := d.Missions.GetMission(ctx, in.TenantID, in.MissionID)
	if err != nil {
		return domain.MissionProgress{}, err
	}
	if !mission.Active {
		return domain.MissionProgress{}, fmt.Errorf("%w: mission inactive", domain.ErrInvalidArgument)
	}
	if _, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID); err != nil {
		return domain.MissionProgress{}, err
	}

	now := d.now()
	prog, err := d.Missions.GetProgress(ctx, in.TenantID, in.AccountID, in.MissionID)
	if err != nil {
		prog = domain.MissionProgress{
			ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID, MissionID: in.MissionID,
			Progress: 0, Status: domain.MissionActive, UpdatedAt: now,
		}
	}
	if prog.Status == domain.MissionCompleted {
		return prog, nil
	}

	prog.Progress += in.Delta
	prog.UpdatedAt = now
	if prog.Progress >= mission.TargetCount {
		return d.completeMissionProgress(ctx, mission, prog)
	}
	if err := d.Missions.UpsertProgress(ctx, prog); err != nil {
		return domain.MissionProgress{}, err
	}
	return prog, nil
}

// CompleteMissionInput force-completes a mission.
type CompleteMissionInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
	MissionID uuid.UUID
}

// CompleteMission marks mission complete, grants points, unlocks linked achievement.
func (d *Deps) CompleteMission(ctx context.Context, in CompleteMissionInput) (domain.MissionProgress, error) {
	mission, err := d.Missions.GetMission(ctx, in.TenantID, in.MissionID)
	if err != nil {
		return domain.MissionProgress{}, err
	}
	now := d.now()
	prog, err := d.Missions.GetProgress(ctx, in.TenantID, in.AccountID, in.MissionID)
	if err != nil {
		prog = domain.MissionProgress{
			ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID, MissionID: in.MissionID,
			Progress: mission.TargetCount, Status: domain.MissionActive, UpdatedAt: now,
		}
	}
	if prog.Status == domain.MissionCompleted {
		return prog, nil
	}
	prog.Progress = mission.TargetCount
	return d.completeMissionProgress(ctx, mission, prog)
}

func (d *Deps) completeMissionProgress(ctx context.Context, mission domain.Mission, prog domain.MissionProgress) (domain.MissionProgress, error) {
	now := d.now()
	prog.Status = domain.MissionCompleted
	prog.CompletedAt = &now
	prog.UpdatedAt = now
	if err := d.Missions.UpsertProgress(ctx, prog); err != nil {
		return domain.MissionProgress{}, err
	}

	acct, err := d.Accounts.GetAccount(ctx, prog.TenantID, prog.AccountID)
	if err != nil {
		return domain.MissionProgress{}, err
	}
	if mission.RewardPoints > 0 {
		acct, _, err = d.grantPoints(ctx, acct, mission.RewardPoints, domain.PointGrant,
			"mission:"+mission.Code, "mission-complete:"+prog.ID.String(), nil)
		if err != nil {
			return domain.MissionProgress{}, err
		}
	}
	d.emit(ctx, acct, domain.EventMissionCompleted, map[string]any{
		"missionId": mission.ID.String(), "code": mission.Code,
	})
	if mission.Achievement != "" {
		_, _ = d.UnlockAchievement(ctx, UnlockAchievementInput{
			TenantID: prog.TenantID, AccountID: prog.AccountID, Code: mission.Achievement,
		})
	}
	return prog, nil
}
