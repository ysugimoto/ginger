package command

import (
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mattn/go-tty"
	"github.com/ysugimoto/go-args"

	"github.com/ysugimoto/ginger/config"
	"github.com/ysugimoto/ginger/input"
	"github.com/ysugimoto/ginger/internal/colors"
	"github.com/ysugimoto/ginger/logger"
	"github.com/ysugimoto/ginger/request"
)

const scheduleExpressionInquiry = `[Schedule expression]
You can use schedule expression like crontab with specific format for AWS:

Cron exression: "cron(0, 10, * * ? *)" executes 10:00 am (UTC) every day
Rate exression: "rate(1 hour)" executes every hour.

Note that the CloudWatchEvents schedules time as UTC, so you need to consider your timezone.
For example, if you want to run 10:00 am (JST), cron becomes "cron(0, 1, * * ? *)" (-9 hours).

See in detail: https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html`

const (
	SCHEDULE_CREATE = "create"
	SCHEDULE_DELETE = "delete"
	SCHEDULE_DEPLOY = "deploy"
	SCHEDULE_LIST   = "list"
	SCHEDULE_ATTACH = "attach"
	SCHEDULE_HELP   = "help"
)

// Schduler is the struct of AWS CloudWatchEvents management command.
// This struct will be dispatched on "ginger schedule" subcommand.
type Scheduler struct {
	Command
	log *logger.Logger
}

func NewSchduler() *Scheduler {
	return &Scheduler{
		log: logger.WithNamespace("ginger.schdule"),
	}
}

// Show function command help.
func (s *Scheduler) Help() string {
	return COMMAND_HEADER + `
schedule - (AWS CloudWatchEvents) management command.

Usage:
  $ ginger schedule|sc [operation] [options]

Operation:
  create : Create new scheduler
  delete : Delete scheduler
  deploy : Deploy scheduler
  attach : Attach scheduler to function
  list   : List schedulers
  help   : Show this help

Options:
  -n, --name : [all] Scheduler name
`
}

// Run the command.
func (s *Scheduler) Run(ctx *args.Context) {
	c := config.Load()
	if !c.Exists() {
		s.log.Error("Configuration file could not load. Run `ginger init` before.")
		return
	}
	var err error
	defer func() {
		if err != nil {
			s.log.Error(err.Error())
			debugTrace(err)
		}
		c.Write()
	}()

	switch ctx.At(1) {
	case SCHEDULER_CREATE:
		err = s.createScheduler(c, ctx)
	case SCHEDULER_DELETE:
		err = s.deleteScheduler(c, ctx)
	case SCHEDULER_DEPLOY:
		err = NewDeploy().deployScheduler(c, ctx)
	case SCHEDULER_LIST:
		err = s.listScheduler(c, ctx)
	case SCHEDULE_ATTACH:
		err = s.attachScheduler(c, ctx)
	default:
		fmt.Println(s.Help())
	}
}

func (s *Scheduler) writeConfig(c *config.Config, sc *entity.Scheduler) error {
	fileName := filepath.Join(c.SchedulerPath, fmt.Sprintf("%s.toml", sc.Name))
	fp, _ := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer fp.Close()
	enc := toml.NewEncoder(fp)
	return enc.Encode(fn)
}

// createScheduler creates new scheduler.
func (s *Scheduler) createScheduler(c *config.Config, ctx *args.Context) (err error) {
	name := ctx.String("name")
	if name == "" {
		name = input.String("Type scheduler name")
	}
	if _, err := c.LoadScheduler(name); err == nil {
		return exception("%s already exists", name)
	}

	fmt.Println(colors.Yellow(scheduleExpressionInquiry))
	expression := input.Strting("Input schedule expression")
	if expression == "" {
		return exception("Abort due to empty expression.")
	}

	enable := input.Bool("Do you want to be enable scheudle immediately?")
	sc := &entity.Scheduler{
		Name:       name,
		Enable:     enable,
		Expression: expression,
	}
	if err := s.writeConfig(sc); err != nil {
		return exception(err.Error())
	}

	s.log.Infof("Scheduler created. To manage its configuration, edit schdulers/%s.toml.\n", name)
	return nil
}

// deleteScheduler deletes stage.
func (s *Scheduler) deleteScheduler(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		return exception("Scheduler name didn't supplied. Run with --name option.")
	}
	sc, err := c.LoadScheduler(name)
	if err != nil {
		return exception("Scheduler not defined.")
	}

	if sc.Arn != "" {
		if !ctx.Has("force") && !input.Bool("Also deletes from AWS CloudWatchEvents. Are you sure?") {
			s.log.Warn("Abort.")
			return nil
		}
		cw := request.NewCloudWatchRequest(c)
		if err := cw.DeleteSchedule(sc); err != nil {
			return nil
		}
	}

	if err := c.DeleteScheduler(name); err != nil {
		return exception(err.Error())
	}
	s.log.Info("Scheduler deleted successfully.")
	return nil
}

// listScheduler shows registered stages.
func (s *Scheduler) listScheduler(c *config.Config, ctx *args.Context) error {
	scs, _ := c.LoadAllSchedulers()
	t, err := tty.Open()
	if err != nil {
		return exception("Couldn't open tty")
	}
	defer t.Close()
	w, _, err := t.Size()
	if err != nil {
		return exception("Couldn't get tty size")
	}
	line := strings.Repeat("=", w)
	fmt.Println(line)
	fmt.Printf("%-12s %-24s %-12s %-12s %-24s\n", "SchedulerName", "Expression", "Enable", "Functions")
	fmt.Println(line)
	for i, sc := range scs {
		f := ""
		if sc.Functions != nil {
			f = strings.Join(sc.Functions, ",")
		}
		if len(f) > 23 {
			f = f[0:23] + "..."
		}
		e := "disabled"
		if sc.Enable {
			e = "enabled"
		}
		fmt.Printf("%-12s %-24s %-12s %-12s %-24s\n", sc.Name, sc.Expression, e, f)
		if i != len(scs)-1 {
			fmt.Println(strings.Repeat("-", w))
		}
	}
	return nil
}

func (s *Scheduler) attachScheduler(c *config.Config, ctx *args.Context) error {
	name := ctx.String("name")
	if name == "" {
		name = c.ChooseScheduler()
	}
	fname := c.ChooseFunction()
	if name == "" || fname == "" {
		return exception("Empty input detected. Abort.")
	}
	sc, err := c.LoadScheduler(name)
	if err != nil {
		return exception(err.Error())
	}

	if _, err := c.LoadFunction(fname); err != nil {
		return exception(err.Error())
	}

	if sc.Functions == nil {
		sc.Functions = make([]string, 1)
	}
	sc.Functions = append(sc.Functions, fname)
	if err := s.writeConfig(sc); err != nil {
		return exception(err.Error())
	}
	s.log.Infof("Schedule %s attached to function %s.\n", name, fname)
	return nil
}
