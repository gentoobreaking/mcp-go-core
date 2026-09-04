package config

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a config file for changes and auto-reloads.
type Watcher struct {
	cfg     *Config
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
	done    chan struct{}
	logger  func(format string, args ...any)
	stopOnce sync.Once
}
// NewWatcher creates a file watcher that auto-reloads the config on save.
// The config must already be loaded once.
func NewWatcher(cfg *Config, logger func(format string, args ...any)) (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = func(format string, args ...any) {}
	}

	return &Watcher{
		cfg:     cfg,
		watcher: w,
		stopCh:  make(chan struct{}),
		done:    make(chan struct{}),
		logger:  logger,
	}, nil
}

// Start begins watching the config file. Must be called after Load().
func (w *Watcher) Start() error {
	if err := w.watcher.Add(w.cfg.path); err != nil {
		return err
	}

	go w.run()
	return nil
}

// run is the event loop. Debounces rapid writes (common on macOS/Windows).
func (w *Watcher) run() {
	defer close(w.done)
	var debounce *time.Timer

	for {
		select {
		case <-w.stopCh:
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(200*time.Millisecond, func() {
					if err := w.cfg.Load(); err != nil {
						w.logger("config reload failed (keeping previous): %v", err)
					} else {
						w.logger("config reloaded successfully at %v", w.cfg.LastLoaded())
					}
				})
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger("config watch error: %v", err)
		}
	}
}

// Stop stops the watcher and waits for the event loop to exit.
// Safe to call multiple times and even when Start was not called.
func (w *Watcher) Stop() error {
	var err error
	w.stopOnce.Do(func() {
		close(w.stopCh)
		select {
		case <-w.done:
		case <-time.After(2 * time.Second):
		}
		err = w.watcher.Close()
	})
	return err
}
// Reload forces a config reload (for /config/reload endpoint).
func (w *Watcher) Reload() error {
	return w.cfg.Load()
}
