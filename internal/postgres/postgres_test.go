package postgres

import (
	"context"
	"fmt"
	u "golang-demo/internal/utils"
	cl "golang-demo/pkg/catelog"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
)

// Statcllient implements the athena Statcllient interface for mocking purposes.
type Statcllient struct {
	CountFn     func(string, float64, []string)
	GaugeFn     func(string, float64, []string)
	HistogramFn func(string, float64, []string)
	HandlerFn   func() http.Handler
}

// Count calls the Statcllient's CountFn.
func (cl *Statcllient) Count(name string, incBy float64, labels []string) {
	cl.CountFn(name, incBy, labels)
}

// Gauge calls the Statcllient's GaugeFn.
func (cl *Statcllient) Gauge(name string, value float64, labels []string) {
	cl.GaugeFn(name, value, labels)
}

// Histogram calls the Statcllient's HistogramFn.
func (cl *Statcllient) Histogram(name string, value float64, labels []string) {
	cl.HistogramFn(name, value, labels)
}

// Handler calls the Statcllient's HTTPHandlerFn.
func (cl *Statcllient) Handler() http.Handler {
	return cl.HandlerFn()
}

// NopStatcllient implements the Statcllient interface where all funcions
// are no-ops.
var NopStatcllient = &Statcllient{
	CountFn:     func(string, float64, []string) {},
	GaugeFn:     func(string, float64, []string) {},
	HistogramFn: func(string, float64, []string) {},
	HandlerFn:   func() http.Handler { return nil },
}

func newPostgres(t *testing.T) *Postgres {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort, _ := strconv.Atoi(os.Getenv("POSTGRES_PORT"))

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == 0 {
		dbPort = 2997
	}

	p, err := New(Config{
		DisableSSL: true,
		Host:       dbHost,
		Port:       dbPort,
		Name:       "story_challenges_test",
		Password:   "",
		Username:   "postgres",
	}, NopStatcllient)
	if err != nil {
		t.Fatalf("Unable to create postgres instance: %s", err.Error())
	}
	return p
}

func clearPostgres(p *Postgres, t *testing.T) {
	_, err := p.sqldb.Exec(`
		TRUNCATE TABLE challenge_creatorgroups CAclADE;
		TRUNCATE TABLE challenge_locations CAclADE;
		TRUNCATE TABLE challenge_stories CAclADE;
		TRUNCATE TABLE challenge_stories_archive CAclADE;
		TRUNCATE TABLE challenges CAclADE;
		TRUNCATE TABLE creatorgroup_creators CAclADE;
		TRUNCATE TABLE creatorgroups CAclADE;
		TRUNCATE TABLE ingestion_results CAclADE;
		TRUNCATE TABLE reviews CAclADE;
		TRUNCATE TABLE content_guidelines CAclADE;
	`)
	if err != nil {
		t.Fatalf("Unable to clear postgres: %s", err.Error())
	}
}

type TestingType interface {
	Create(ctx context.Context, db *sqlx.DB) error
}

func createTestInstance(ctx context.Context, t *testing.T, db *sqlx.DB, v TestingType) {
	t.Helper()
	err := v.Create(ctx, db)
	if err != nil {
		t.Fatalf("error creating %T: %s", v, err.Error())
	}
}

func createTestChallenge(ctx context.Context, p *Postgres, t *testing.T) TestingChallenge {
	t.Helper()
	cID := u.NewUUID(t)
	oID := null.NewString(u.NewUUID(t), true)
	challenge := TestingChallenge{
		ID:             cID,
		Label:          fmt.Sprintf("Test %s Label", cID),
		CreatedAt:      time.Time{},
		OrganizationID: oID,
		Status:         cl.ChallengeStatusOpen,
		CreatorLimit:   10,
		ChallengeType:  cl.ChallengeTypeStandard,
	}
	createTestInstance(ctx, t, p.sqldb, &challenge)
	return challenge
}

func createCustomTestChallenge(ctx context.Context, p *Postgres, t *testing.T, c cl.Challenge) TestingChallenge {
	t.Helper()
	challenge := TestingChallenge{
		ID:             c.ID,
		Label:          c.Label,
		CreatedAt:      c.CreatedAt,
		OrganizationID: c.OrganizationID,
		Status:         c.Status,
		CreatorLimit:   c.CreatorLimit,
		ChallengeType:  c.ChallengeType,
	}
	createTestInstance(ctx, t, p.sqldb, &challenge)
	return challenge
}

func assignTestStoryToChallenge(ctx context.Context, p *Postgres, t *testing.T, storyID, challengeID string, publisherSlug string) cl.ChallengeStory {
	t.Helper()

	currentTime := time.Now()
	storyAssignment := cl.ChallengeStory{
		StoryID:       null.NewString(storyID, true),
		ChallengeID:   challengeID,
		Approved:      false,
		PublisherSlug: publisherSlug,
		AssignedAt:    null.NewTime(currentTime, true),
		SubmittedAt:   null.NewTime(currentTime, true),
		ApprovedAt:    null.Time{},
		UpdatedAt:     currentTime,
		State:         cl.ChallengeStoryStateSubmitted,
	}

	req1 := cl.AssignStoryRequest{
		ChallengeStory: storyAssignment,
	}

	res, err := p.AssignStory(ctx, req1)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return res.ChallengeStory
}

func assignApproveTestStoryToChallenge(ctx context.Context, p *Postgres, t *testing.T, storyID, challengeID string, publisherSlug string) cl.ChallengeStory {
	t.Helper()
	currentTime := time.Now()

	storyAssignment := cl.ChallengeStory{
		StoryID:       null.NewString(storyID, true),
		ChallengeID:   challengeID,
		Approved:      true,
		PublisherSlug: publisherSlug,
		AssignedAt:    null.NewTime(currentTime, true),
		SubmittedAt:   null.NewTime(currentTime, true),
		ApprovedAt:    null.Time{},
		UpdatedAt:     currentTime,
		State:         cl.ChallengeStoryStateSubmitted,
	}

	aReq := cl.AssignStoryRequest{
		ChallengeStory: storyAssignment,
	}

	_, err := p.AssignStory(ctx, aReq)
	if err != nil {
		t.Fatalf(err.Error())
	}

	appReq := cl.ApproveStoryRequest{
		ChallengeStory: storyAssignment,
	}
	appRes, err := p.ApproveStory(ctx, appReq)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return appRes.ChallengeStory
}
