package executor

import (
	"strings"

	"github.com/AbGuthrie/goquery/api/models"
	"github.com/AbGuthrie/goquery/commands"
	"github.com/AbGuthrie/goquery/utils"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// Executor is the main execution engine for goquery.
type Executor struct {
	driver   *types.API
	history  *types.History
	commands map[string]types.Command // TODO maybe make this a pointer to an  interface?
	aliased  *types.Aliases
	logger   log.Logger
}

type Opt func(*Executor) error

func WithLogger(logger log.Logger) Opt {
	return func(ex *Executor) error {
		ex.logger = logger
		return nil
	}
}

func WithHistory(h *models.History) Opt {
	return func(ex *Executor) error {
		ex.history = h
		return nil
	}
}

func New(driver models.GoQueryAPI, opts ...Opt) (*Executor, error) {
	ex := &Executor{
		driver: driver,
		logger: log.NewNopLogger(),
	}

	for _, opt := range opts {
		if err := opt(ex); err != nil {
			return nil, errors.Wrap(err, "creating new executor")
		}
	}

	return ex, nil
}

// PromptRun acts as a shim between prompt and Run. It handles the run errors and returns nothing
func (ex *Executor) PromptRun(input string) {
	if err := ex.Run(input); err != nil {
		levels.Info(ex.logger).Log(
			"msg", "Got error running command",
			"err", err,
		)
	}
	return
}

// Run parses input as a command and runs it.
func (ex *Executor) Run(input string) error {
	// Separate command and arguments
	input = strings.TrimSpace(input)
	args := strings.Split(input, " ")
	if len(args) == 0 {
		return
	}

	// After we're done executing, write history. This is in a
	// defer block, so alias resolution can skip writing history.
	skipHistory := false
	defer func() {
		if skipHistory || ex.history == nil {
			return
		}

		// Write history entry. Non fatal error
		if err := ex.history.Update(input); err != nil {
			level.Info(ex.logger).Log(
				"msg", "Failed to write history file",
				"err", err,
			)
		}
	}()

	// Is this command in the command map? If so, let's run it!
	// This is an early return, as no command means we should try
	// to de-alias it
	if command, ok := commands.CommandMap[args[0]]; ok {
		err := command.Execute(input)
		return errors.Wrap(err, "executing command")
	}

	// Command not found, was this command aliased?
	alias, found := ex.aliases.Resolve(args[0])
	if !found {
		return errors.Errorf("No such command or alias: %s", args[0])
	}

	// TODO: seph This isn't well integrated in this model
	realizedCommand, err := utils.InterpolateArguments(input, alias.Command)
	if err != nil {
		return errors.Wrap(err, "alias interpoation")
	}

	// Run the parsed and interpolated alias through executor again
	skipHistory = false
	return ex.Run(realizedCommand)
}

func (ex *Executor) Completer(in prompt.Document) []prompt.Suggest {
	prompts := []prompt.Suggest{}
	command := strings.Split(in.CurrentLine(), " ")[0]

	// Nothing has been typed at the prompt
	if command == "" {
		return prompts
	}

	// FIXME: not implemented
	return prompts
}
