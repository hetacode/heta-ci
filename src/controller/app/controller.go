package app

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/hetacode/heta-ci/controller/db"
	"github.com/hetacode/heta-ci/controller/enums"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/structs"
	"github.com/xo/dburl"
)

type Controller struct {
	Repositories     []utils.Repository
	Builds           map[string]*utils.PipelineBuild
	BuildLastCommits utils.BuildLastCommits
	pipelines        []*structs.Pipeline
	agents           []*utils.Agent // list of free agents

	ReturnAgentCh         chan *utils.Agent            // after finished job agent back via channel
	AskAgentCh            chan string                  // build id as parameter
	buildsAgentResponseCh map[string]chan *utils.Agent // channels collection for each build - via these channels are sending free agents to execute jobs
	addAgentCh            chan *utils.Agent
	removeAgentCh         chan *utils.Agent
}

func NewController(addAgentCh, removeAgentCh chan *utils.Agent) *Controller {
	c := &Controller{
		Builds:                make(map[string]*utils.PipelineBuild),
		BuildLastCommits:      make(utils.BuildLastCommits),
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

func (c *Controller) RegisterBuild(build *utils.PipelineBuild, repositoryHash string, commitHash string) {
	c.Builds[build.ID] = build
	c.buildsAgentResponseCh[build.ID] = build.AgentResponseChan

	// TODO: move creating connection to the some common place
	conn, err := dburl.Open("pgsql://postgres:postgrespass@localhost/heta-ci?sslmode=disable")
	if err != nil {
		log.Panicf("open connection to database failed %s", err)
	}

	pipelineBytes, _ := json.Marshal(build.Pipeline)
	uid, _ := uuid.FromString(build.ID)
	dbBuild := &db.Build{
		UID:            uid,
		RepositoryHash: repositoryHash,
		CommitHash:     commitHash,
		PipelineJSON:   string(pipelineBytes),
		ResultStatus:   string(enums.BuildStatusNone),
		CreatedAt:      time.Now().Unix(),
	}
	dbBuild.Insert(conn)
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
