package plugin

import (
	"mylife-home-common/log"
	"mylife-home-core-library/definitions"

	"github.com/jsuar/go-cron-descriptor/pkg/crondescriptor"
	"github.com/robfig/cron/v3"
)

var logger = log.CreateLogger("mylife:home:core:plugins:logic-timers")

// @Plugin(usage="logic")
type Scheduler struct {

	// @Config(description="Definition cron du scheduler - https://crontab.guru/ - https://en.wikipedia.org/wiki/Cron")
	Cron string

	// @State()
	Enabled definitions.State[bool]

	// @State(description="Sortie du sheduler")
	Trigger definitions.State[bool]

	// @State(description="Description du planning dans un format lisible pour un humain")
	Schedule definitions.State[string]

	// @State(description="Timestamp JS avant le prochain déclenchement")
	NextDate definitions.State[float64]

	executor   definitions.Executor
	cronEngine *cron.Cron
	cronJobId  cron.EntryID // 0 = no job
}

func (component *Scheduler) Init(runtime definitions.Runtime) error {
	component.Enabled.Set(true)

	component.executor = runtime.NewExecutor()
	component.cronEngine = cron.New()

	if err := component.setupScheduler(); err != nil {
		logger.WithError(err).Error("Error initializing scheduler")
		return nil
	}

	component.cronEngine.Start()
	component.refreshNextDate()

	logger.Debug("Scheduler starting job")
	return nil
}

func (component *Scheduler) setupScheduler() error {

	desc, err := crondescriptor.NewCronDescriptor(component.Cron)
	if err != nil {
		return err
	}

	value, err := desc.GetDescription(crondescriptor.Full)
	if err != nil {
		return err
	}

	component.Schedule.Set(*value)

	cronJobId, err := component.cronEngine.AddFunc(component.Cron, component.onTick)
	if err != nil {
		return err
	}

	component.cronJobId = cronJobId
	return nil
}

func (component *Scheduler) Terminate() {
	logger.Debug("Scheduler stopping job")

	ctx := component.cronEngine.Stop()
	<-ctx.Done()

	component.executor.Terminate()
}

// @Action(description="Permet de désactiver le scheduler. Le trigger reste alors à 'false'")
func (component *Scheduler) Disable(arg bool) {
	value := !arg
	component.Enabled.Set(value)
	logger.Debugf("Scheduler changed job state to '%t'", value)
}

func (component *Scheduler) onTick() {
	component.executor.Execute(func() {
		if component.Enabled.Get() {
			logger.Debug("Scheduler trigger")

			component.Trigger.Set(true)
			component.Trigger.Set(false)
		} else {
			logger.Debug("Skipping scheduler trigger (disabled)")
		}

		component.refreshNextDate()
	})
}

func (component *Scheduler) refreshNextDate() {
	// paranoia
	if component.cronJobId == 0 {
		return
	}

	entry := component.cronEngine.Entry(component.cronJobId)
	// paranoia
	if !entry.Valid() {
		return
	}

	// Javascript timestamp
	nextTs := float64(entry.Next.UnixMilli())

	component.NextDate.Set(nextTs)
}
