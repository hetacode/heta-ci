package app

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	intlstructs "github.com/hetacode/heta-ci/controller/structs"

	"github.com/hetacode/heta-ci/controller/db"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/structs"
)

type Controller struct {
	DBRepository     db.DBRepository
	Repositories     []utils.Repository
	Builds           map[string]*utils.PipelineBuild
	BuildLastCommits intlstructs.BuildLastCommits
	pipelines        []*structs.Pipeline
	agents           []*utils.Agent // list of free agents

	ReturnAgentCh         chan *utils.Agent            // after finished job agent back via channel
	AskAgentCh            chan string                  // build id as parameter
	buildsAgentResponseCh map[string]chan *utils.Agent // channels collection for each build - via these channels are sending free agents to execute jobs
	addAgentCh            chan *utils.Agent
	removeAgentCh         chan *utils.Agent
}

func NewController(dbRepository db.DBRepository, addAgentCh, removeAgentCh chan *utils.Agent) *Controller {
	c := &Controller{
		DBRepository:          dbRepository,
		Builds:                make(map[string]*utils.PipelineBuild),
		BuildLastCommits:      make(intlstructs.BuildLastCommits),
		pipelines:             make([]*structs.Pipeline, 0),
		agents:                make([]*utils.Agent, 0),
		buildsAgentResponseCh: make(map[string]chan *utils.Agent),
		AskAgentCh:            make(chan string),
		ReturnAgentCh:         make(chan *utils.Agent),
		addAgentCh:            addAgentCh,
		removeAgentCh:         removeAgentCh,
	}
	go c.agentsManager()

	if err := os.Mkdir(utils.PipelinesDir, 0777); err != nil {
		log.Printf("start controller | warning: create %s directory failed | %s", utils.PipelinesDir, err)
	}
	if err := os.Mkdir(utils.RepositoryDirectory, 0777); err != nil {
		log.Printf("start controller | warning: create %s directory failed | %s", utils.RepositoryDirectory, err)
	}

	return c
}

func (c *Controller) AddPipeline(p *structs.Pipeline) {
	c.pipelines = append(c.pipelines, p)
}

func (c *Controller) RegisterBuild(build *utils.PipelineBuild, repositoryHash string, commitHash string) error {
	if err := c.DBRepository.StoreBuildData(build.ID, build.Pipeline, repositoryHash, commitHash); err != nil {
		return fmt.Errorf("store build data in db failed %s", err)
	}

	c.Builds[build.ID] = build
	c.buildsAgentResponseCh[build.ID] = build.AgentResponseChan

	return nil
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
		case buildID := <-c.AskAgentCh:
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
