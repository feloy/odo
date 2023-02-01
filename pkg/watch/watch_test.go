package watch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/fsnotify/fsnotify"
)

func evaluateChangesHandler(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) ([]string, []string) {
	var changedFiles []string
	var deletedPaths []string

	for _, event := range events {
		for _, file := range fileIgnores {
			// if file is in fileIgnores, don't do anything
			if event.Name == file {
				continue
			}
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
			// On remove/rename, stop watching the resource
			_ = watcher.Remove(event.Name)
			// Append the file to list of deleted files
			deletedPaths = append(deletedPaths, event.Name)
			continue
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove != fsnotify.Remove {
			changedFiles = append(changedFiles, event.Name)
		}
	}
	return changedFiles, deletedPaths
}

func processEventsHandler(ctx context.Context, changedFiles, deletedPaths []string, _ WatchParameters, out io.Writer, componentStatus *ComponentStatus, backo *ExpBackoff) (*time.Duration, error) {
	fmt.Fprintf(out, "changedFiles %s deletedPaths %s\n", changedFiles, deletedPaths)
	return nil, nil
}

type fakeWatcher struct{}

func (o fakeWatcher) Stop() {
}

func (o fakeWatcher) ResultChan() <-chan watch.Event {
	return make(chan watch.Event, 1)
}

type fakePodWatcher struct {
	ch chan watch.Event
}

func newFakePodWatcher() fakePodWatcher {
	return fakePodWatcher{
		ch: make(chan watch.Event, 1),
	}
}

func (o fakePodWatcher) Stop() {
}

func (o fakePodWatcher) ResultChan() <-chan watch.Event {
	return o.ch
}

func (o fakePodWatcher) sendPodReady() {
	o.ch <- watch.Event{
		Type: watch.Added,
		Object: &corev1.Pod{
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	}
}

func Test_eventWatcher(t *testing.T) {
	type args struct {
		parameters WatchParameters
	}
	tests := []struct {
		name          string
		args          args
		wantOut       string
		wantErr       bool
		watcherEvents []fsnotify.Event
		watcherError  error
	}{
		{
			name: "Case 1: Multiple events, no errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       " ✓  Pod is Running\nPushing files...\n\nchangedFiles [file1 file2] deletedPaths []\n",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  nil,
		},
		{
			name: "Case 2: Multiple events, one error",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       " ✓  Pod is Running\n",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  fmt.Errorf("error"),
		},
		{
			name: "Case 3: Delete file, no error",
			args: args{
				parameters: WatchParameters{FileIgnores: []string{"file1"}},
			},
			wantOut:       " ✓  Pod is Running\nPushing files...\n\nchangedFiles [] deletedPaths [file1 file2]\n",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Remove}, {Name: "file2", Op: fsnotify.Rename}},
			watcherError:  nil,
		},
		{
			name: "Case 4: Only errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: nil,
			watcherError:  fmt.Errorf("error1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, _ := fsnotify.NewWatcher()
			fileWatcher, _ := fsnotify.NewWatcher()
			podWatcher := newFakePodWatcher()
			var cancel context.CancelFunc
			ctx, cancel := context.WithCancel(context.Background())
			out := &bytes.Buffer{}

			go func() {
				podWatcher.sendPodReady()
				for _, event := range tt.watcherEvents {
					watcher.Events <- event
				}

				if tt.watcherError != nil {
					watcher.Errors <- tt.watcherError
				}
				<-time.After(500 * time.Millisecond)
				cancel()
			}()

			componentStatus := ComponentStatus{
				State: StateReady,
			}

			o := WatchClient{
				sourcesWatcher:    watcher,
				deploymentWatcher: fakeWatcher{},
				podWatcher:        podWatcher,
				warningsWatcher:   fakeWatcher{},
				devfileWatcher:    fileWatcher,
				keyWatcher:        make(chan byte),
			}
			err := o.eventWatcher(ctx, tt.args.parameters, out, evaluateChangesHandler, processEventsHandler, componentStatus)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("eventWatcher() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
