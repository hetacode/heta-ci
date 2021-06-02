package jobs

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hetacode/heta-ci/controller/utils"
	"github.com/hetacode/heta-ci/structs"
	"gopkg.in/yaml.v2"
)

type RepositoryPeriodicJob struct {
	controller *utils.Controller
	lastRun    int64
	interval   time.Duration
	isRunning  bool
}

func NewRepositoryPeriodicJob(interval time.Duration, ctrl *utils.Controller) *RepositoryPeriodicJob {
	j := &RepositoryPeriodicJob{
		controller: ctrl,
		lastRun:    0,
		interval:   interval,
		isRunning:  false,
	}

	return j
}

func (j *RepositoryPeriodicJob) Init() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			period := time.Now().Unix() - j.interval.Milliseconds()
			if j.lastRun < period {
				continue
			}
			if j.isRunning {
				continue
			}

			j.lastRun = time.Now().Unix()
			j.Run()
		}
	}()
}

func (j *RepositoryPeriodicJob) Run() {
	for _, r := range j.controller.Repositories {
		fmt.Printf("repo: %s default branch: %s \n", r.Url, r.DefaultBranch)
		rc, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:           r.Url,
			ReferenceName: plumbing.NewBranchReferenceName(r.DefaultBranch),
		})
		if err != nil {
			log.Printf("repo %s clone failed %s", r.Url, err)
			continue
		}
		ref, _ := rc.Head()
		commit, _ := rc.CommitObject(ref.Hash())
		tree, err := commit.Tree()
		if err != nil {
			log.Printf("get repo tree failed %s", err)
			continue
		}

		pf, err := tree.File(".heta-ci/pipeline.yaml")
		if err != nil {
			fmt.Printf("read pipeline file failed err: %s", err)
			continue
		}
		c, _ := pf.Contents()
		var pipeline *structs.Pipeline
		if err := yaml.Unmarshal([]byte(c), &pipeline); err != nil {
			fmt.Printf("unmarshal pipeline file failed err %s", err)
			return
		}

		refs, err := rc.Storer.IterReferences()
		if err != nil {
			fmt.Printf("err IterReferences: %s", err)
			return
		}

		//  Get all remote branches
		branches := make([]string, 0)
		refs.ForEach(func(ref *plumbing.Reference) error {
			if ref.Name().IsRemote() {
				branches = append(branches, ref.Name().Short())
			}
			return nil
		})
		for _, r := range pipeline.RunOn {
			switch r.Type {
			case structs.RunOnBranch:
				findAndRunBranches(r.On, branches)
			}
		}
	}
}

func findAndRunBranches(runOnPattern string, branches []string) error {
	checkPattern := regexp.MustCompile(runOnPattern)
	for _, b := range branches {
		correctPattern := checkPattern.MatchString(b)
		fmt.Printf("check '%s' pattern: %s - %t \n", runOnPattern, b, correctPattern)
		if correctPattern {
			// TODO: run build
			// build should check last processed commit and run newest version of code
		}
	}

	return nil
}
