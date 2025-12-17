package rbcmd

type PwdCommand struct{}

func (c *PwdCommand) Help() string {
	return "Print the current working directory: pwd"
}

func (c *PwdCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	return *ctx.CWD, "", nil
}

