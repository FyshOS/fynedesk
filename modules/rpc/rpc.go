package rpc

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"

	"fyshos.com/tyde"
)

// LaunchInput passes the input string through each LaunchSuggestionModule
// and executes the first match. Search module is ignored in this context.
func LaunchInput(input string) (string, error) {
	desk := tyde.Instance()
	if desk == nil {
		return "", errors.New("desktop not running")
	}

	for _, m := range desk.Modules() {
		suggest, ok := m.(tyde.LaunchSuggestionModule)
		if !ok {
			continue
		}
		items := suggest.LaunchSuggestions(input)
		if len(items) == 0 {
			continue
		}

		// Search module is a catch-all fallbacks which we don't want.
		if strings.Contains(strings.ToLower(m.Metadata().Name), "search") {
			continue
		}

		item := items[0]
		done := make(chan struct{})
		fyne.Do(func() {
			item.Launch()
			close(done)
		})
		<-done
		return item.Title(), nil
	}

	return "", fmt.Errorf("no match for %q", input)
}

// Run performs a line of text, as typed at tyde_ctl: the command named by its
// first word if there is one, otherwise the first launch suggestion the input
// matches. Returns an error if nothing matches.
//
// This must parse the command parts, a caller that already holds a Command should
// use RunCommand instead.
func Run(input string) (string, error) {
	if tyde.Instance() == nil {
		return "", errors.New("desktop not running")
	}

	trimmed := strings.TrimSpace(input)
	name, args, _ := strings.Cut(trimmed, " ")

	for _, cmd := range Commands() {
		if strings.EqualFold(name, cmd.Name) {
			return RunCommand(cmd, args)
		}
	}
	return LaunchInput(trimmed)
}

// RunCommand performs a command from Commands, on the UI thread where module
// code expects to be.
func RunCommand(cmd tyde.Command, args string) (string, error) {
	var reply string
	var err error
	fyne.DoAndWait(func() {
		reply, err = cmd.Run(strings.TrimSpace(args))
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", cmd.Name, err)
	}
	return reply, nil
}

// Commands lists everything Run understands: the desktop's own commands plus
// those each loaded module offers.
func Commands() []tyde.Command {
	cmds := append([]tyde.Command(nil), builtins...)

	desk := tyde.Instance()
	if desk == nil {
		return cmds
	}
	for _, m := range desk.Modules() {
		provider, ok := m.(tyde.CommandModule)
		if !ok {
			continue
		}
		for _, cmd := range provider.Commands() {
			if cmd.Run == nil {
				// Nothing to run means nothing to offer. Caught here, where the
				// commands are gathered, rather than by every caller.
				fyne.LogError("Module "+m.Metadata().Name+" describes command "+
					cmd.Name+" with no action", nil)
				continue
			}
			cmds = append(cmds, cmd)
		}
	}
	return cmds
}

// builtins are the commands the desktop itself answers.
var builtins = []tyde.Command{
	{
		Name:        "welcome",
		Description: "Open the FyshOS welcome and first-run setup window",
		Run: func(string) (string, error) {
			welcomer, ok := tyde.Instance().(interface{ ShowWelcome() })
			if !ok {
				return "", errors.New("this desktop has no welcome window")
			}
			welcomer.ShowWelcome()
			return "Welcome window", nil
		},
	},
	{
		// The assistant is a window the desktop opens, not something it can ask
		// the assistant module to describe: an AI that offers itself a command to
		// open itself is no use to anyone.
		Name:        "fathom",
		Description: "Open the Fathom AI assistant chat window",
		Run: func(string) (string, error) {
			desk := tyde.Instance()
			if desk == nil {
				return "", errors.New("desktop not running")
			}

			for _, m := range desk.Modules() {
				if chat, ok := m.(interface{ ShowChat() }); ok {
					chat.ShowChat()
					return "Fathom AI", nil
				}
			}
			return "", errors.New("the AI assistant module is not enabled")
		},
	},
	{
		Name:        "restart",
		Description: "Restart the desktop, keeping the session and open apps",
		Run: func(string) (string, error) {
			if os.Getenv("FYNE_DESK_RUNNER") == "" {
				return "", errors.New("tyde was not launched with tyde_runner, cannot restart")
			}
			os.Exit(5)
			return "", nil
		},
	},
}

var rpcMeta = tyde.ModuleMetadata{
	Name:        "Remote Control",
	NewInstance: newRPC,
}

// SocketPath returns the Unix socket path used for RPC communication.
func SocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir != "" {
		return filepath.Join(dir, "tyde.sock")
	}
	return "/tmp/tyde-" + strconv.Itoa(os.Getuid()) + ".sock"
}

type rpcModule struct {
	listener net.Listener
}

func (r *rpcModule) Metadata() tyde.ModuleMetadata {
	return rpcMeta
}

func (r *rpcModule) Destroy() {
	if r.listener != nil {
		r.listener.Close()
	}
	os.Remove(SocketPath())
}

// Service is the RPC service exposed over the Unix socket.
type Service struct{}

// Launch runs a desktop built-in, or passes the input string through the
// launch suggestion modules and executes the first match.
func (s *Service) Launch(input string, reply *string) error {
	title, err := Run(input)
	if err != nil {
		return err
	}
	*reply = title
	return nil
}

// ListCommands returns every command Launch understands, so a caller can offer
// them without knowing which modules are loaded.
func (s *Service) ListCommands(_ string, reply *[]tyde.Command) error {
	if tyde.Instance() == nil {
		return errors.New("desktop not running")
	}
	*reply = Commands()
	return nil
}

func newRPC() tyde.Module {
	sock := SocketPath()
	os.Remove(sock) // clean up stale socket

	mod := &rpcModule{}
	srv := rpc.NewServer()
	srv.Register(&Service{})

	ln, err := net.Listen("unix", sock)
	if err != nil {
		fyne.LogError("RPC: failed to listen on "+sock, err)
		return &rpcModule{}
	} else {
		log.Println("RPC: Listening on " + sock)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go srv.ServeConn(conn)
		}
	}()

	mod.listener = ln
	return mod
}
