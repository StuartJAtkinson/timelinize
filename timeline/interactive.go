/*
	Timelinize
	Copyright (c) 2013 Matthew Holt

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published
	by the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package timeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const interactiveTextPreviewLimit = 64 * 1024

func (p *processor) interactiveGraph(ctx context.Context, root *Graph, opts *InteractiveImport) error {
	if root == nil {
		return errors.New("interactive import received an empty graph")
	}
	p.assignGraphIDs(root)
	defer func() {
		if err := os.RemoveAll(p.tempGraphFolder()); err != nil {
			p.log.Warn("cleaning up interactive graph files", zap.Error(err))
		}
	}()
	if err := p.downloadGraphDataFiles(ctx, root); err != nil {
		return err
	}

	interactiveGraph := &InteractiveGraph{
		Graph:     root,
		Submitted: make(chan struct{}),
	}
	p.log.Info("graph ready", zap.String("graph_id", root.ProcessingID))

	select {
	case opts.Graphs <- interactiveGraph:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-interactiveGraph.Submitted:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *processor) assignGraphIDs(g *Graph) {
	visited := make(map[*Graph]struct{})
	var assign func(*Graph)
	assign = func(g *Graph) {
		if g == nil {
			return
		}
		if _, ok := visited[g]; ok {
			return
		}
		visited[g] = struct{}{}
		if g.ProcessingID == "" {
			g.ProcessingID = uuid.New().String()
		}
		for _, edge := range g.Edges {
			assign(edge.From)
			assign(edge.To)
		}
	}
	assign(g)
}

func (p *processor) downloadGraphDataFiles(ctx context.Context, root *Graph) error {
	visited := make(map[*Graph]struct{})
	var download func(*Graph) error
	download = func(g *Graph) error {
		if g == nil {
			return nil
		}
		if _, ok := visited[g]; ok {
			return nil
		}
		visited[g] = struct{}{}
		if err := ctx.Err(); err != nil {
			return err
		}

		var dataReader io.ReadCloser
		var dataFilename string
		switch {
		case g.Item != nil && g.Item.Content.Data != nil:
			dataFilename = g.Item.Content.Filename
			reader, err := g.Item.Content.Data(ctx)
			if err != nil {
				return fmt.Errorf("opening item data reader: %w", err)
			}
			if reader == nil {
				return errors.New("item data reader is nil")
			}
			dataReader = reader
		case g.Entity != nil && g.Entity.NewPicture != nil:
			reader, err := g.Entity.NewPicture(ctx)
			if err != nil {
				return fmt.Errorf("opening entity picture reader: %w", err)
			}
			if reader == nil {
				return errors.New("entity picture reader is nil")
			}
			dataReader = reader
		}

		if dataReader != nil {
			if g.Item != nil {
				g.Item.Content.DataAvailable = true
			}
			file, err := p.createInteractiveGraphDataFile(g)
			if err != nil {
				_ = dataReader.Close()
				return fmt.Errorf("creating interactive graph data file: %w", err)
			}
			dataReader, err = p.fillMediaTypeOfInteractiveGraph(g, dataReader, dataFilename)
			if err != nil {
				_ = file.Close()
				if !errors.Is(err, io.EOF) {
					_ = dataReader.Close()
					_ = os.Remove(file.Name())
					return fmt.Errorf("detecting interactive graph media type: %w", err)
				}
			} else {
				if _, err := copyWithContext(ctx, file, dataReader); err != nil {
					_ = dataReader.Close()
					_ = file.Close()
					_ = os.Remove(file.Name())
					return fmt.Errorf("staging interactive graph data: %w", err)
				}
				if err := file.Sync(); err != nil {
					_ = dataReader.Close()
					_ = file.Close()
					_ = os.Remove(file.Name())
					return fmt.Errorf("syncing interactive graph data: %w", err)
				}
				if err := file.Close(); err != nil {
					_ = dataReader.Close()
					_ = os.Remove(file.Name())
					return fmt.Errorf("closing interactive graph data: %w", err)
				}
			}
			if err := dataReader.Close(); err != nil {
				return fmt.Errorf("closing interactive graph data reader: %w", err)
			}
			if g.Item != nil && strings.HasPrefix(g.Item.Content.MediaType, "text/") {
				if err := p.loadInteractiveTextPreview(g); err != nil {
					return err
				}
			}
		}

		if g.Item != nil && g.Item.Owner.NewPicture != nil {
			reader, err := g.Item.Owner.NewPicture(ctx)
			if err != nil {
				return fmt.Errorf("opening item owner picture reader: %w", err)
			}
			if reader == nil {
				return errors.New("item owner picture reader is nil")
			}
			filename := InteractiveGraphOwnerPictureFilePath(p.tl.ID(), p.ij.job.ID(), g.ProcessingID)
			if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
				_ = reader.Close()
				return fmt.Errorf("creating interactive owner picture directory: %w", err)
			}
			file, err := os.Create(filename)
			if err != nil {
				_ = reader.Close()
				return fmt.Errorf("creating interactive owner picture file: %w", err)
			}
			if _, err := copyWithContext(ctx, file, reader); err != nil {
				_ = reader.Close()
				_ = file.Close()
				_ = os.Remove(filename)
				return fmt.Errorf("staging interactive owner picture: %w", err)
			}
			if err := file.Sync(); err != nil {
				_ = reader.Close()
				_ = file.Close()
				_ = os.Remove(filename)
				return fmt.Errorf("syncing interactive owner picture: %w", err)
			}
			if err := file.Close(); err != nil {
				_ = reader.Close()
				_ = os.Remove(filename)
				return fmt.Errorf("closing interactive owner picture: %w", err)
			}
			if err := reader.Close(); err != nil {
				return fmt.Errorf("closing interactive owner picture reader: %w", err)
			}
		}

		for _, edge := range g.Edges {
			if err := download(edge.From); err != nil {
				return err
			}
			if err := download(edge.To); err != nil {
				return err
			}
		}
		return nil
	}
	return download(root)
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 32*1024)
	var total int64
	emptyReads := 0
	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}
		n, readErr := src.Read(buf)
		if n == 0 && readErr == nil {
			emptyReads++
			if emptyReads >= 100 {
				return total, io.ErrNoProgress
			}
		} else {
			emptyReads = 0
		}
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			total += int64(written)
			if writeErr != nil {
				return total, writeErr
			}
			if written != n {
				return total, io.ErrShortWrite
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return total, nil
			}
			return total, readErr
		}
	}
}

func (p *processor) loadInteractiveTextPreview(g *Graph) error {
	filename := InteractiveGraphDataFilePath(p.tl.ID(), p.ij.job.ID(), g.ProcessingID)
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening interactive text preview: %w", err)
	}
	defer file.Close()

	preview, err := io.ReadAll(io.LimitReader(file, interactiveTextPreviewLimit+1))
	if err != nil {
		return fmt.Errorf("reading interactive text preview: %w", err)
	}
	if len(preview) > interactiveTextPreviewLimit {
		preview = append(preview[:interactiveTextPreviewLimit], []byte("\n[preview truncated]")...)
	}
	g.Item.Content.PreviewText = string(preview)
	return nil
}

func restoreInteractiveGraphData(original, submitted *Graph, repoID uuid.UUID, jobID uint64, restoreData bool) error {
	type graphPair struct {
		original  *Graph
		submitted *Graph
	}
	visited := make(map[graphPair]struct{})
	var restore func(*Graph, *Graph) error
	restore = func(original, submitted *Graph) error {
		if original == nil || submitted == nil {
			if original == submitted {
				return nil
			}
			return errors.New("graph structure changed during review")
		}
		pair := graphPair{original: original, submitted: submitted}
		if _, ok := visited[pair]; ok {
			return nil
		}
		visited[pair] = struct{}{}
		if original.ProcessingID == "" || original.ProcessingID != submitted.ProcessingID {
			return errors.New("graph node ID changed during review")
		}
		if (original.Item == nil) != (submitted.Item == nil) || (original.Entity == nil) != (submitted.Entity == nil) {
			return errors.New("graph node type changed during review")
		}
		if restoreData && original.Item != nil && original.Item.Content.Data != nil {
			filename := InteractiveGraphDataFilePath(repoID, jobID, original.ProcessingID)
			if !FileExists(filename) {
				return fmt.Errorf("staged item data is missing: %s", filename)
			}
			submitted.Item.Content.Data = func(ctx context.Context) (io.ReadCloser, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				return os.Open(filename)
			}
		}
		if restoreData && original.Item != nil && original.Item.Owner.NewPicture != nil {
			filename := InteractiveGraphOwnerPictureFilePath(repoID, jobID, original.ProcessingID)
			if !FileExists(filename) {
				return fmt.Errorf("staged item owner picture is missing: %s", filename)
			}
			submitted.Item.Owner.NewPicture = func(ctx context.Context) (io.ReadCloser, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				return os.Open(filename)
			}
		}
		if restoreData && original.Entity != nil && original.Entity.NewPicture != nil {
			filename := InteractiveGraphDataFilePath(repoID, jobID, original.ProcessingID)
			if !FileExists(filename) {
				return fmt.Errorf("staged entity picture is missing: %s", filename)
			}
			submitted.Entity.NewPicture = func(ctx context.Context) (io.ReadCloser, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				return os.Open(filename)
			}
		}
		if len(original.Edges) != len(submitted.Edges) {
			return errors.New("graph relationships changed during review")
		}
		for i := range original.Edges {
			if err := restore(original.Edges[i].From, submitted.Edges[i].From); err != nil {
				return err
			}
			if err := restore(original.Edges[i].To, submitted.Edges[i].To); err != nil {
				return err
			}
		}
		return nil
	}
	return restore(original, submitted)
}

func countGraphItems(root *Graph) int {
	visited := make(map[*Graph]struct{})
	var count int
	var walk func(*Graph)
	walk = func(g *Graph) {
		if g == nil {
			return
		}
		if _, ok := visited[g]; ok {
			return
		}
		visited[g] = struct{}{}
		if g.Item != nil {
			count++
		}
		for _, edge := range g.Edges {
			walk(edge.From)
			walk(edge.To)
		}
	}
	walk(root)
	return count
}

func (p *processor) fillMediaTypeOfInteractiveGraph(g *Graph, dataReader io.ReadCloser, dataFilename string) (io.ReadCloser, error) {
	// if there is no item data file, or it already has a content-type, no-op
	if g.Item == nil || g.Item.Content.MediaType != "" {
		return dataReader, nil
	}

	const peekSize = 1024
	buf := make([]byte, peekSize)

	n, err := io.ReadFull(dataReader, buf)

	// ignore ErrUnexpectedEOF, it just means the content is short;
	// return if there's EOF (caller must handle) -- means no content;
	// obviously return if there's any other error too
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return dataReader, err
	}

	// in case the content was short, use only written portion of buffer
	peekedBytes := buf[:n]

	g.Item.Content.MediaType = detectContentType(peekedBytes, dataFilename)

	// replace the reader with one that re-reads the buffered bytes before resuming reading the source stream
	dataReader = multiReadCloser{
		Reader: io.MultiReader(bytes.NewReader(peekedBytes), dataReader),
		Closer: dataReader,
	}

	return dataReader, nil
}

type multiReadCloser struct {
	io.Reader
	io.Closer
}

func (p *processor) createInteractiveGraphDataFile(g *Graph) (*os.File, error) {
	// We store interactive graph data files, temporarily while the user is
	// interacting with the graph, somewhat deep in the system temp folder.
	// It's in a system temp folder because import jobs are not typically
	// portable; especially starting on one system and continuing on another,
	// though I guess we could simply change the path to be something within
	// the timeline if desired. Still, this seems more proper at least for now.
	tmpFilePath := InteractiveGraphDataFilePath(p.tl.ID(), p.ij.job.ID(), g.ProcessingID)

	// ensure folder tree exists or we're gonna have a bad time
	if err := os.MkdirAll(filepath.Dir(tmpFilePath), 0700); err != nil {
		return nil, err
	}

	return os.Create(tmpFilePath)
}

func InteractiveGraphDataFilePath(repoID uuid.UUID, jobID uint64, graphID string) string {
	return filepath.Join(tempGraphFolder(repoID, jobID), graphID+".graph.data")
}

func InteractiveGraphOwnerPictureFilePath(repoID uuid.UUID, jobID uint64, graphID string) string {
	return filepath.Join(tempGraphFolder(repoID, jobID), graphID+".owner.picture")
}

func (p *processor) tempGraphFolder() string {
	return tempGraphFolder(p.tl.ID(), p.ij.job.ID())
}

func tempGraphFolder(repoID uuid.UUID, jobID uint64) string {
	return filepath.Join(appTempDir(), fmt.Sprintf("tl-%s", repoID), fmt.Sprintf("job-%d", jobID))
}

func appTempDir() string {
	return filepath.Join(os.TempDir(), "timelinize")
}
