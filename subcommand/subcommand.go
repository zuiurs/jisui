package subcommand

type Command struct {
	Name string
	Run  func(args []string) error
}

var Commands []*Command
