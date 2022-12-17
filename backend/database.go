package backend

import (
	"fmt"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Start the local database
func (a *Surrealist) StartDatabase(user string, pass string, port uint32, driver string, storage string) {
	if a.isServing {
		a.StopDatabase()
	}

	args := []string{
		"start",
		"--bind", fmt.Sprintf("0.0.0.0:%d", port),
		"--user", user,
		"--pass", pass,
	}

	switch driver {
	case "memory":
		args = append(args, "memory")
	case "file":
		args = append(args, "file://"+storage)
	case "tikv":
		args = append(args, "tikv://"+storage)
	}

	go func() {
		a.isServing = true

		defer func() {
			a.isServing = false
			a.serverHandle = nil

			runtime.EventsEmit(a.ctx, "database:stop")
			runtime.LogInfo(a.ctx, "Local database stopped")
		}()

		cmd := exec.Command("surreal", args...)

		if err := cmd.Start(); err != nil {
			return
		}

		a.serverHandle = cmd.Process

		runtime.EventsEmit(a.ctx, "database:start")
		runtime.LogInfo(a.ctx, fmt.Sprintf("Local database started with args: %v", args))

		cmd.Wait()
	}()
}

// Stop the local database
func (a *Surrealist) StopDatabase() {
	if !a.isServing {
		return
	}

	// We should probably not kill but send a SIGINT, right?
	// But this works for now.
	err := a.serverHandle.Kill()

	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Failed to kill local database: %v", err))
	}
}