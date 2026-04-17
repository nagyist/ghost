package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/timescale/ghost/internal/api"
	"github.com/timescale/ghost/internal/util"
)

type WaitForDatabaseArgs struct {
	Client      api.ClientWithResponsesInterface
	ProjectID   string
	DatabaseRef string
}

func WaitForDatabase(ctx context.Context, args WaitForDatabaseArgs) error {
	const timeout = 10 * time.Minute

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				return ExitWithCode(ExitTimeout, err)
			default:
				return err
			}
		case <-ticker.C:
			resp, err := args.Client.GetDatabaseWithResponse(ctx, args.ProjectID, args.DatabaseRef)
			if err != nil {
				return fmt.Errorf("failed to fetch database status: %w", err)
			}

			switch resp.StatusCode() {
			case 200:
				if resp.JSON200 == nil {
					return errors.New("no response body returned from API")
				}

				// Check returned database status
				if resp.JSON200.Status == api.DatabaseStatusRunning {
					return nil
				}
			case 404:
				// Can happen if user deletes database while it's still provisioning
				return errors.New("database not found")
			case 500:
				return errors.New("internal server error")
			default:
				return fmt.Errorf("received unexpected %s", resp.Status())
			}
		}
	}
}

// WaitForDatabaseWithProgress waits for a database to be ready, showing an
// animated spinner if the writer is a terminal, or plain text otherwise.
func WaitForDatabaseWithProgress(ctx context.Context, out io.Writer, args WaitForDatabaseArgs) error {
	if !util.IsTerminal(out) {
		return WaitForDatabase(ctx, args)
	}

	model := waitModel{
		spinner: spinner.New(spinner.WithSpinner(spinner.Ellipsis)),
		ctx:     ctx,
		args:    args,
	}

	p := tea.NewProgram(
		model,
		tea.WithInput(nil),
		tea.WithOutput(out),
		tea.WithoutSignalHandler(),
	)

	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("error rendering to terminal: %w", err)
	}

	return result.(waitModel).err
}

type waitModel struct {
	spinner spinner.Model
	ctx     context.Context
	args    WaitForDatabaseArgs
	done    bool
	err     error
}

type waitResultMsg struct {
	err error
}

func (m waitModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			return waitResultMsg{err: WaitForDatabase(m.ctx, m.args)}
		},
	)
}

func (m waitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case waitResultMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m waitModel) View() tea.View {
	switch {
	case m.err != nil:
		return tea.NewView("")
	case m.done:
		return tea.NewView("Database is ready\n")
	default:
		return tea.NewView("Waiting for database" + m.spinner.View())
	}
}
