package jobs

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/go-uuid"
	"github.com/hetacode/heta-ci/commons"
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
		for _, ro := range pipeline.RunOn {
			switch ro.Type {
			case structs.RunOnBranch:
				j.findAndRunBranches(pipeline, &r, ro.On, branches)
			}
		}
	}
}

func (j *RepositoryPeriodicJob) findAndRunBranches(pipeline *structs.Pipeline, repository *utils.Repository, runOnPattern string, branches []string) error {
	checkPattern := regexp.MustCompile(runOnPattern)
	for _, b := range branches {
		correctPattern := checkPattern.MatchString(b)
		fmt.Printf("repoID: %s check '%s' pattern: %s - %t \n", repository.ID, runOnPattern, b, correctPattern)
		if correctPattern {
			// TODO: run build
			// build should check last processed commit and run newest version of code
			if err := j.prepareBuildPipeline(pipeline, repository, structs.RunOnBranch, b); err != nil {
				fmt.Printf("prepareBuildPipeline err %s\n", err)
			}

		}
	}

	return nil
}

func (j *RepositoryPeriodicJob) prepareBuildPipeline(pipeline *structs.Pipeline, repository *utils.Repository, runOnType structs.RunOnType, runOnValue string) error {
	// TODO: fetch repo and check commits diffs
	var repoBytes []byte
	lastCommit := j.controller.BuildLastCommits.Get(repository.ID, runOnType, runOnValue)
	if lastCommit == nil {
		rc, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:           repository.Url,
			ReferenceName: plumbing.NewBranchReferenceName(strings.TrimPrefix(runOnValue, "origin/")),
		})
		if err != nil {
			return fmt.Errorf("repo %s clone branch %s failed %s", repository.Url, runOnValue, err)
		}
		ref, _ := rc.Head()
		commit, _ := rc.CommitObject(ref.Hash())
		tree, err := commit.Tree()
		if err != nil {
			return fmt.Errorf("get repo tree failed %s", err)
		}

		convertIter := &convertGitFileIterToZipFileIter{
			iter: tree.Files(),
		}

		repoBytes, err = commons.ArchiveFiles(convertIter)
		if err != nil {
			return fmt.Errorf("archive repo failed %s", err)
		}
	} else {
		panic("prepareBuildPipeline - unimplemented diffs flow")
	}
	repositoryArchiveID, _ := uuid.GenerateUUID()
	pipeline.RepositoryArchiveID = repositoryArchiveID
	filePath := fmt.Sprintf("%s/%s.zip", utils.RepositoryDirectory, repositoryArchiveID)
	os.WriteFile(filePath, repoBytes, 0777)
	return nil
	// w := utils.NewPipelineBuild(pipeline, j.controller.AskAgentCh)
	// j.controller.RegisterBuild(w)
	// w.Run()
}

type convertGitFileIterToZipFileIter struct {
	iter *object.FileIter
}

func (c *convertGitFileIterToZipFileIter) ForEach(callback func(file *commons.FileData) error) error {
	err := c.iter.ForEach(func(o *object.File) error {
		r, err := o.Reader()
		if err != nil {
			return fmt.Errorf("file reader err %s", err)
		}
		data := &commons.FileData{
			Path:   o.Name,
			Reader: r,
		}

		if err := callback(data); err != nil {
			return err
		}
		return nil
	})
	return err
}
