package utils

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/hetacode/heta-ci/structs"
)

type Controller struct {
	Builds    map[string]*PipelineBuild
	pipelines []*structs.Pipeline
	agents    []*Agent // list of free agents

	ReturnAgentCh         chan *Agent            // after finished job agent back via channel
	buildsAgentResponseCh map[string]chan *Agent // channels collection for each build - via these channels are sending free agents to execute jobs
	askAgentCh            chan string            // build id as parameter
	addAgentCh            chan *Agent
	removeAgentCh         chan *Agent
}

func NewController(addAgentCh, removeAgentCh chan *Agent) *Controller {
	c := &Controller{
		Builds:                make(map[string]*PipelineBuild),
		pipelines:             make([]*structs.Pipeline, 0),
		agents:                make([]*Agent, 0),
		buildsAgentResponseCh: make(map[string]chan *Agent),
		askAgentCh:            make(chan string),
		ReturnAgentCh:         make(chan *Agent),
		addAgentCh:            addAgentCh,
		removeAgentCh:         removeAgentCh,
	}
	go c.agentsManager()

	if err := os.Mkdir(PipelinesDir, 0777); err != nil {
		log.Printf("start controller | warning: create %s directory failed | %s", PipelinesDir, err)
	}

	return c
}

func (c *Controller) AddPipeline(p *structs.Pipeline) {
	c.pipelines = append(c.pipelines, p)
}

func (c *Controller) Execute() {
	// TODO: a correct way - it should iterate through git repositories
	for _, p := range c.pipelines {
		w := NewPipelineBuild(p, c.askAgentCh)
		c.Builds[w.ID] = w
		c.buildsAgentResponseCh[w.ID] = w.AgentResponseChan
		w.Run()
	}

}

func (c *Controller) agentsManager() {
	builds := make([]string, 0)
	var wg sync.WaitGroup

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if len(builds) > 0 {
				wg.Wait()
				wg.Add(1)
				if len(c.agents) > 0 {
					buildID := builds[0]
					builds = append(builds[1:])
					agent := c.agents[0]
					c.agents = append(c.agents[1:])
					log.Printf("agent %s has been assign to build: %s", agent.ID, buildID)
					c.buildsAgentResponseCh[buildID] <- agent
				}
				wg.Done()
			}
		}
	}()

	for {
		select {
		case buildID := <-c.askAgentCh:
			builds = append(builds, buildID)
		case agent := <-c.addAgentCh:
			wg.Wait()
			wg.Add(1)
			c.agents = append(c.agents, agent)
			log.Printf("added agent %s ", agent.ID)
			wg.Done()
		case agent := <-c.removeAgentCh:
			wg.Wait()
			wg.Add(1)
			for i, a := range c.agents {
				if a.ID == agent.ID {
					c.agents = append(c.agents[:i], c.agents[i+1:]...)
					break
				}
			}
			log.Printf("removed agent %s", agent.ID)
			wg.Done()
		case agent := <-c.ReturnAgentCh:
			wg.Wait()
			wg.Add(1)
			c.agents = append(c.agents, agent)
			log.Printf("agent %s is free again", agent.ID)
			wg.Done()
		}
	}
}
