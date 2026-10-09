package socialmediaarchiver_test

import (
	"context"
	"os"
	"testing"

	"github.com/StuartJAtkinson/timelinize/datasources/socialmediaarchiver"
	"github.com/StuartJAtkinson/timelinize/internal/testhelpers"
	"github.com/StuartJAtkinson/timelinize/timeline"
)

func dirEntry() timeline.DirEntry {
	return timeline.DirEntry{FS: os.DirFS("testdata/fixtures"), Filename: "output"}
}

func TestRecognize(t *testing.T) {
	rec, err := new(socialmediaarchiver.Importer).Recognize(context.Background(), dirEntry(), timeline.RecognizeParams{})
	if err != nil || rec.Confidence != 1 {
		t.Fatalf("archive not recognised: confidence=%v err=%v", rec.Confidence, err)
	}
	other := timeline.DirEntry{FS: os.DirFS("testdata/fixtures/output"), Filename: "images"}
	if rec, _ := new(socialmediaarchiver.Importer).Recognize(context.Background(), other, timeline.RecognizeParams{}); rec.Confidence != 0 {
		t.Errorf("non-archive folder recognised with confidence %v", rec.Confidence)
	}
}

func TestFileImport(t *testing.T) {
	pipeline := make(chan *timeline.Graph, 20)
	if err := new(socialmediaarchiver.Importer).FileImport(context.Background(), dirEntry(), timeline.ImportParams{Pipeline: pipeline}); err != nil {
		t.Fatalf("import: %v", err)
	}
	close(pipeline)

	got := map[string]*timeline.Graph{}
	for g := range pipeline {
		got[g.Item.ID] = g
	}
	for _, uid := range []string{"bluesky:p1", "mastodon:m1", "reddit:r1", "twitter:t1", "youtube_community:y1", "rss:s1", "facebook:f1"} {
		if got[uid] == nil {
			t.Errorf("missing item %s", uid)
		}
	}
	if len(got) != 7 {
		t.Fatalf("want 7 items (comments sidecar and non-item JSON skipped), got %d", len(got))
	}

	g := got["bluesky:p1"]
	if g.Item.Timestamp.Format("2006-01-02T15:04:05Z07:00") != "2026-05-01T10:00:00Z" {
		t.Errorf("timestamp = %v", g.Item.Timestamp)
	}
	if g.Item.Metadata["URL"] != "https://bsky.app/profile/alice.bsky.social/post/p1" {
		t.Errorf("original URL not kept: %v", g.Item.Metadata["URL"])
	}
	testhelpers.ValidateItemData(t, "", []byte("hello from bluesky"), g.Item.Content, "text")

	owner := g.Item.Owner
	if owner.Name != "Alice" || len(owner.Attributes) == 0 || owner.Attributes[0].Name != "bluesky_id" || owner.Attributes[0].Value != "did:plc:alice" || !owner.Attributes[0].Identity {
		t.Errorf("author entity wrong: %+v", owner)
	}
	if a := got["reddit:r1"].Item.Owner.Attributes; len(a) == 0 || a[0].Value != "carol" {
		t.Errorf("author without id should fall back to name: %+v", a)
	}

	// one local media file attached; the undownloaded one (local_path null) is skipped
	if len(g.Edges) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(g.Edges))
	}
	att := g.Edges[0].To.Item
	if att.Content.Filename != "abc.png" {
		t.Errorf("attachment filename = %q", att.Content.Filename)
	}
	testhelpers.ValidateItemData(t, "", []byte("PNGDATA"), att.Content, "attachment")
}
