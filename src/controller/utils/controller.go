package utils

import (
	"log"
	"sync"

	"github.com/hetacode/heta-ci/structs"
)

type Controller struct {
	Builds    map[string]*PipelineBuild
	pipelines []*structs.Pipeline
	agents    []*Agent // list of free agents

	buildsAgentResponseCh map[string]chan *Agent // channels collection for each build - via these channels are sending free agents to execute jobs
	askAgentCh            chan string            // build id as parameter
	returnAgentCh         chan *Agent            // after finished job agent back via channel
}

func NewController() *Controller {
	c := &Controller{}
	go c.agentsManager()

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
		go w.Run()
	}
}

func (c *Controller) agentsManager() {
	builds := make([]string, 0)
	var wg sync.WaitGroup
	for {
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

		select {
		case buildID := <-c.askAgentCh:
			builds = append(builds, buildID)
		case agent := <-c.returnAgentCh:
			wg.Wait()
			wg.Add(1)
			c.agents = append(c.agents, agent)
			log.Printf("agent %s is free again", agent.ID)
			wg.Done()
		}
	}
}
