package socialmediaarchiver_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/StuartJAtkinson/timelinize/datasources/twitter"
	"github.com/StuartJAtkinson/timelinize/timeline"
)

// TestDedupAgainstTwitterArchive imports the same tweet through an official
// Twitter archive and through socialMediaArchiver and expects one item.
func TestDedupAgainstTwitterArchive(t *testing.T) {
	ctx := context.Background()
	tl, err := timeline.Create(ctx, filepath.Join(t.TempDir(), "repo"))
	if err != nil {
		t.Fatalf("creating timeline: %v", err)
	}
	defer tl.Close()
	tl.SetObfuscationFunc(func() (timeline.ObfuscationOptions, bool) { return timeline.ObfuscationOptions{}, false }) // normally set by the app

	abs := func(p string) string {
		a, err := filepath.Abs(filepath.Join("testdata", "fixtures", p))
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	jobID, err := tl.CreateJob(&timeline.ImportJob{Plan: timeline.ImportPlan{Files: []timeline.FileImport{
		{DataSourceName: "twitter", DataSourceOptions: []byte(`{}`), Filenames: []string{abs("twitter-archive")}},
		{DataSourceName: "socialmediaarchiver", Filenames: []string{abs("output")}},
	}}}, time.Time{}, 0, 0, 0)
	if err != nil {
		t.Fatalf("creating import job: %v", err)
	}

	deadline := time.Now().Add(60 * time.Second)
	for {
		jobs, err := tl.GetJobs(ctx, []uint64{jobID}, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(jobs) == 1 && (jobs[0].State == timeline.JobSucceeded || jobs[0].State == timeline.JobFailed) {
			if jobs[0].State == timeline.JobFailed {
				msg := ""
				if jobs[0].Message != nil {
					msg = *jobs[0].Message
				}
				t.Fatalf("import job failed: %s", msg)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("import job did not finish in time")
		}
		time.Sleep(100 * time.Millisecond)
	}

	res, err := tl.Search(ctx, timeline.ItemSearchParams{Classification: []string{"social"}, Limit: -1})
	if err != nil {
		t.Fatal(err)
	}
	// 7 archiver posts (one per platform) + 1 archive tweet that is the same post as the archiver's twitter item
	if len(res.Items) != 7 {
		t.Fatalf("want 7 social items (tweet merged, not duplicated), got %d", len(res.Items))
	}
}
