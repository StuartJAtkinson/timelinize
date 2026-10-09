// Package socialmediaarchiver imports the output folder of socialMediaArchiver,
// which normalises Bluesky, Mastodon, Reddit, X, YouTube community posts, RSS
// and Facebook Pages into one item shape written as <output>/<source>/<id>.json.
package socialmediaarchiver

import (
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/StuartJAtkinson/timelinize/timeline"
	"go.uber.org/zap"
)

func init() {
	err := timeline.RegisterDataSource(timeline.DataSource{
		Name:            "socialmediaarchiver",
		Title:           "socialMediaArchiver",
		Icon:            "socialmediaarchiver.svg",
		Description:     "Import posts saved by socialMediaArchiver (Bluesky, Mastodon, Reddit, X, YouTube, RSS, Facebook Pages)",
		NewFileImporter: func() timeline.FileImporter { return new(Importer) },
	})
	if err != nil {
		timeline.Log.Fatal("registering data source", zap.Error(err))
	}
}

// item mirrors core/models.py Item.to_dict() in socialMediaArchiver.
type item struct {
	UID       string `json:"uid"`
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	URL       string `json:"url"`
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Author    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"author"`
	Media []struct {
		URL       string `json:"url"`
		LocalPath string `json:"local_path"`
		MediaType string `json:"media_type"`
	} `json:"media"`
}

// valid reports whether this decoded as a socialMediaArchiver item rather than some other JSON.
func (it item) valid() bool {
	return it.ID != "" && it.UID == it.Source+":"+it.ID
}

// Importer imports a socialMediaArchiver output folder.
type Importer struct{}

// Recognize returns whether the input is a socialMediaArchiver output folder.
func (Importer) Recognize(_ context.Context, dirEntry timeline.DirEntry, _ timeline.RecognizeParams) (timeline.Recognition, error) {
	found := false
	_ = walkItems(dirEntry, func(string, item) error {
		found = true
		return fs.SkipAll
	})
	if found {
		return timeline.Recognition{Confidence: 1}, nil
	}
	return timeline.Recognition{}, nil
}

// FileImport imports every item in the folder.
func (Importer) FileImport(ctx context.Context, dirEntry timeline.DirEntry, params timeline.ImportParams) error {
	return walkItems(dirEntry, func(_ string, it item) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		params.Pipeline <- graph(dirEntry, it)
		return nil
	})
}

// walkItems calls fn for every <source>/<id>.json item one level below the root.
func walkItems(dirEntry timeline.DirEntry, fn func(string, item) error) error {
	sources, err := dirEntry.ReadDir(".")
	if err != nil {
		return err
	}
	for _, src := range sources {
		if !src.IsDir() || strings.HasPrefix(src.Name(), ".") {
			continue
		}
		files, err := dirEntry.ReadDir(src.Name())
		if err != nil {
			return err
		}
		for _, f := range files {
			name := f.Name()
			if f.IsDir() || path.Ext(name) != ".json" || strings.HasSuffix(name, "_comments.json") {
				continue
			}
			p := path.Join(src.Name(), name)
			it, ok := readItem(dirEntry, p)
			if !ok {
				continue
			}
			if err := fn(p, it); err != nil {
				if err == fs.SkipAll {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

func readItem(dirEntry timeline.DirEntry, p string) (item, bool) {
	f, err := dirEntry.Open(p)
	if err != nil {
		return item{}, false
	}
	defer f.Close()
	var it item
	if json.NewDecoder(f).Decode(&it) != nil || !it.valid() {
		return item{}, false
	}
	return it, true
}

func graph(dirEntry timeline.DirEntry, it item) *timeline.Graph {
	ts, _ := time.Parse(time.RFC3339, it.Timestamp)
	owner := author(it)

	meta := timeline.Metadata{"URL": it.URL, "Platform": it.Source}
	if it.Title != "" {
		meta["Title"] = it.Title
	}
	if it.Target != "" {
		meta["Target"] = it.Target
	}

	g := &timeline.Graph{Item: &timeline.Item{
		ID:             it.UID,
		Classification: timeline.ClassSocial,
		Timestamp:      ts,
		Owner:          owner,
		Content:        timeline.ItemData{Data: timeline.StringData(it.Text)},
		Metadata:       meta,
	}}

	for _, m := range it.Media {
		local := resolve(dirEntry, m.LocalPath)
		if local == "" {
			continue // ponytail: not downloaded by the archiver; fetch m.URL here if remote-only media matters
		}
		g.ToItem(timeline.RelAttachment, &timeline.Item{
			Classification: timeline.ClassMedia,
			Timestamp:      ts,
			Owner:          owner,
			Content: timeline.ItemData{
				Filename: path.Base(local),
				Data: func(context.Context) (io.ReadCloser, error) {
					return dirEntry.Open(local)
				},
			},
			Metadata: timeline.Metadata{"URL": m.URL},
		})
	}
	return g
}

// author is one entity per platform account, keyed on "<source>_id".
func author(it item) timeline.Entity {
	id := it.Author.ID
	if id == "" {
		id = it.Author.Name
	}
	if id == "" {
		return timeline.Entity{}
	}
	ent := timeline.Entity{
		Name: it.Author.Name,
		Attributes: []timeline.Attribute{{
			Name:     it.Source + "_id",
			Value:    id,
			Identity: true,
		}},
	}
	if it.Author.URL != "" {
		ent.Metadata = timeline.Metadata{"Profile URL": it.Author.URL}
	}
	return ent
}

// resolve maps a local_path (relative to wherever the archiver ran, e.g.
// `output\images\x.jpg`, or absolute) onto a path inside the output folder by
// dropping leading components until the file is found. Returns "" if absent.
func resolve(dirEntry timeline.DirEntry, localPath string) string {
	p := strings.Trim(strings.ReplaceAll(localPath, `\`, "/"), "/")
	for p != "" {
		if fs.ValidPath(p) && dirEntry.FileExists(p) {
			return p
		}
		_, rest, ok := strings.Cut(p, "/")
		if !ok {
			return ""
		}
		p = rest
	}
	return ""
}
