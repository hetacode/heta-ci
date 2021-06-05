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
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v2"
)

type RepositoryPeriodicJob struct {
	controller *utils.Controller
	lastRun    int64
	cron       string
	isRunning  bool
}

func NewRepositoryPeriodicJob(cron string, ctrl *utils.Controller) *RepositoryPeriodicJob {
	j := &RepositoryPeriodicJob{
		controller: ctrl,
		lastRun:    0,
		cron:       cron,
		isRunning:  false,
	}

	return j
}

func (j *RepositoryPeriodicJob) Init() {
	j.lastRun = time.Now().Unix()
	c := cron.New()
	c.AddFunc(j.cron, func() {
		if j.isRunning {
			return
		}

		j.lastRun = time.Now().Unix()
		j.run()
	})
	c.Start()
}

func (j *RepositoryPeriodicJob) run() {
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
		isCorrectPattern := checkPattern.MatchString(b)
		fmt.Printf("repoID: %s check '%s' pattern: %s - %t \n", repository.ID, runOnPattern, b, isCorrectPattern)
		if isCorrectPattern {
			if err := j.prepareBuildPipeline(pipeline, repository, structs.RunOnBranch, b); err != nil {
				fmt.Printf("prepareBuildPipeline err %s\n", err)
			}

		}
	}

	return nil
}

func (j *RepositoryPeriodicJob) prepareBuildPipeline(pipeline *structs.Pipeline, repository *utils.Repository, runOnType structs.RunOnType, runOnValue string) error {
	var repoBytes []byte
	lastCommitHash := j.controller.BuildLastCommits.Get(repository.ID, runOnType, runOnValue)
	if lastCommitHash == nil {
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

		repoBytes, err = j.archiveRepoAndSaveLastCommit(tree, ref.Hash().String(), repository.ID, runOnType, runOnValue)
		if err != nil {
			return err
		}
	} else {
		rc, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
			URL:           repository.Url,
			ReferenceName: plumbing.NewBranchReferenceName(strings.TrimPrefix(runOnValue, "origin/")),
		})
		if err != nil {
			return fmt.Errorf("repo %s clone branch %s failed %s", repository.Url, runOnValue, err)
		}
		ref, _ := rc.Head()

		// If head is the same as last commit - no changes
		if ref.Hash().String() == *lastCommitHash {
			return nil
		}

		headCommit, err := rc.CommitObject(ref.Hash())
		if err != nil {
			return fmt.Errorf("cannot fetch commit for %s hash (head) err %s", *lastCommitHash, err)
		}
		headTree, _ := headCommit.Tree()

		oldCommit, err := rc.CommitObject(plumbing.NewHash(*lastCommitHash))
		if err != nil {
			return fmt.Errorf("cannot fetch commit for %s hash err %s", *lastCommitHash, err)
		}
		oldTree, err := oldCommit.Tree()
		if err != nil {
			return fmt.Errorf("cannot find old tree from commit object for %s hash err %s", *lastCommitHash, err)
		}
		changes, err := headTree.Diff(oldTree)
		if err != nil {
			return fmt.Errorf("diff failed between %s (head) - %s (last commit) err %s", ref.Hash().String(), *lastCommitHash, err)
		}
		if changes.Len() == 0 {
			return nil // no changes
		}

		log.Printf("start job for: %s repo | %s branch | %s last hash | %s head hash", repository.Url, runOnValue, *lastCommitHash, ref.Hash().String())
		repoBytes, err = j.archiveRepoAndSaveLastCommit(headTree, ref.Hash().String(), repository.ID, runOnType, runOnValue)
		if err != nil {
			return err
		}
	}
	repositoryArchiveID, _ := uuid.GenerateUUID()
	pipeline.RepositoryArchiveID = repositoryArchiveID
	filePath := fmt.Sprintf("%s/%s.zip", utils.RepositoryDirectory, repositoryArchiveID)
	if err := os.WriteFile(filePath, repoBytes, 0777); err != nil {
		return fmt.Errorf("save repository archive failed | path %s err %s", filePath, err)
	}

	w := utils.NewPipelineBuild(pipeline, j.controller.AskAgentCh)
	j.controller.RegisterBuild(w)
	go w.Run()

	fmt.Println("end processeing " + runOnValue)
	return nil
}

func (j *RepositoryPeriodicJob) archiveRepoAndSaveLastCommit(tree *object.Tree, commitSha, repositoryID string, runOnType structs.RunOnType, runOnValue string) ([]byte, error) {
	convertIter := &convertGitFileIterToZipFileIter{
		iter: tree.Files(),
	}

	repoBytes, err := commons.ArchiveFiles(convertIter)
	if err != nil {
		return nil, fmt.Errorf("archive repo failed %s", err)
	}

	j.controller.BuildLastCommits.Add(repositoryID, runOnType, runOnValue, commitSha)

	return repoBytes, nil
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
