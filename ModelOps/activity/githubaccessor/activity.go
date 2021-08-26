/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package githubaccessor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git"
	. "github.com/go-git/go-git/_examples"
	"github.com/go-git/go-git/plumbing"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/plumbing/transport/http"
	"github.com/go-git/go-git/storage/memory"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/logger"
)

var log = logger.GetLogger("tibco-modelops-githubaccessor")

var initialized bool = false

const (
	sFolder = "Folder"

	iCommand       = "Command"
	iUsername      = "Username"
	iPassword      = "Password"
	iProject       = "Project"
	iGitRepository = "GitRepository"

	oDataOut = "DataOut"

	CMD_CheckInfo = "checkInfo"
	CMD_Clone     = "clone"
	CMD_Commit    = "commit"
	CMD_Push      = "push"
)

type GitHubAccessorActivity struct {
	metadata    *activity.Metadata
	mux         sync.Mutex
	workFolders map[string]string
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aGitHubAccessorActivity := &GitHubAccessorActivity{
		metadata:    metadata,
		workFolders: make(map[string]string),
	}

	return aGitHubAccessorActivity
}

func (a *GitHubAccessorActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *GitHubAccessorActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[GitHubAccessorActivity:Eval] entering ........ ")
	defer log.Info("[GitHubAccessorActivity:Eval] Exit ........ ")

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}

	url, ok := context.GetInput(iGitRepository).(string)
	if !ok {
		return false, errors.New("Invalid url ... ")
	}

	username, ok := context.GetInput(iUsername).(string)
	if !ok {
		return false, errors.New("Invalid username ... ")
	}

	password, ok := context.GetInput(iPassword).(string)
	if !ok {
		return false, errors.New("Invalid password ... ")
	}

	project, ok := context.GetInput(iProject).(string)
	if !ok {
		return false, errors.New("Invalid project ... ")
	}

	workFolder, err := a.getWorkfolder(context)

	log.Info("command : ", command)
	log.Info("url : ", url)
	log.Info("username : ", username)
	log.Info("password : ", password)
	log.Info("workFolder : ", workFolder)

	var data map[string]interface{}
	switch command {
	case CMD_Clone:
		{
			data, err = a.clone(url, project, workFolder, username, password)
		}
	case CMD_CheckInfo:
		{
			data, err = a.checkInfo(url, username, password)
		}
	case CMD_Commit:
		{
			data, err = a.commit(url, project, workFolder, username, password)
		}
	case CMD_Push:
		{
			data, err = a.push(url, project, workFolder, username, password)
		}
	default:
		{
			err = errors.New(fmt.Sprintf("Illegal command : %s", command))
		}
	}

	if nil == data {
		data = make(map[string]interface{})
	}

	if nil != err {
		data["error"] = err
	}

	context.SetOutput(oDataOut, data)
	log.Info("[GitHubAccessorActivity:Eval] oDataOut = ", data)

	return true, nil
}

func (a *GitHubAccessorActivity) getWorkfolder(context activity.Context) (string, error) {
	myId := ActivityId(context)
	workFolder := a.workFolders[myId]

	if "" == workFolder {
		a.mux.Lock()
		defer a.mux.Unlock()
		workFolder = a.workFolders[myId]
		if "" == workFolder {
			aFolder, ok := context.GetSetting(sFolder)
			if !ok {
				return "", errors.New("Invalid localFolder ... ")
			}
			workFolder = aFolder.(string)
		}
	}
	return workFolder, nil
}

func (a *GitHubAccessorActivity) prepareCheckoutFolder(workFolder string, project string) (string, error) {
	log.Info("(prepareCheckoutFolder) workFolder = ", workFolder, ", project = ", project)
	checkoutFolder := fmt.Sprintf("%s/%s", workFolder, project)

	log.Info("(prepareCheckoutFolder) check stat of checkoutFolder = ", checkoutFolder)
	err := os.RemoveAll(checkoutFolder)
	if nil != err {
		return "", err
	}

	_, err = os.Stat(workFolder)
	if os.IsNotExist(err) {
		log.Warn("workFolder not exists will try to make it")
		err = os.MkdirAll(workFolder, os.ModePerm)
		if nil != err {
			return "", err
		}
	}
	log.Info("(prepareCheckoutFolder) done ")

	return checkoutFolder, err
}

func (a *GitHubAccessorActivity) clone(url string, project string, workFolder string, username string, password string) (map[string]interface{}, error) {

	data := make(map[string]interface{})
	path, err := a.prepareCheckoutFolder(workFolder, project)
	if nil != err {
		return data, err
	}

	// Clone the given repository, creating the remote, the local branches
	// and fetching the objects, exactly as:
	Info("git clone %s %s", url, path)

	r, err := git.PlainClone(path, false, &git.CloneOptions{URL: url})
	CheckIfError(err)

	// Getting the latest commit on the current branch
	Info("git log -1")

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	CheckIfError(err)

	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	CheckIfError(err)
	fmt.Println(commit)

	// List the tree from HEAD
	Info("git ls-tree -r HEAD")

	// ... retrieve the tree from the commit
	tree, err := commit.Tree()
	CheckIfError(err)

	// ... get the files iterator and print the file
	tree.Files().ForEach(func(f *object.File) error {
		fmt.Printf("100644 blob %s    %s\n", f.Hash, f.Name)
		return nil
	})

	// List the history of the repository
	Info("git log --oneline")

	commitIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	CheckIfError(err)

	err = commitIter.ForEach(func(c *object.Commit) error {
		hash := c.Hash.String()
		line := strings.Split(c.Message, "\n")
		fmt.Println(hash[:7], line[0])

		return nil
	})
	CheckIfError(err)

	descLocation := fmt.Sprintf("%s/%s/pipeline/descriptor.json", workFolder, project)
	log.Info("[GitHubAccessorActivity:clone] descLocation : ", descLocation)

	descriptor, err := readFile(descLocation)
	if nil != err {
		return data, err
	}

	data["descriptor"] = descriptor

	log.Info("[GitHubAccessorActivity:clone] Exit ........ ")

	return data, err
}

func (a *GitHubAccessorActivity) checkInfo(url string, username string, password string) (map[string]interface{}, error) {
	log.Info("[GitHubAccessorActivity:checkInfo] entering ........ ")
	defer log.Info("[GitHubAccessorActivity:checkInfo] Exit ........ ")
	// Clones the given repository, creating the remote, the local branches
	// and fetching the objects, everything in memory:
	//Info("git clone https://github.com/src-d/go-siva")
	data := make(map[string]interface{})
	data["exist"] = false

	storage := memory.NewStorage()
	r, err := git.Clone(storage, nil, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL: url,
	})
	if nil != err {
		return data, nil
	}

	iter, _ := storage.IterEncodedObjects(plumbing.TreeObject)
	iter.ForEach(func(obj plumbing.EncodedObject) error {
		log.Info(obj)
		return nil
	})

	// Gets the HEAD history from HEAD, just like this command:
	log.Info("git log")

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	if nil != err {
		return data, nil
	}

	// ... retrieves the commit history
	since := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2020, 7, 30, 0, 0, 0, 0, time.UTC)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Since: &since, Until: &until})
	if nil != err {
		return data, nil
	}

	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		log.Info(c)
		return nil
	})

	data["exist"] = true

	return data, err
}

func (a *GitHubAccessorActivity) commit(url string, project string, workFolder string, username string, password string) (map[string]interface{}, error) {

	// Opens an already existing repository.
	path, err := a.prepareCheckoutFolder(workFolder, project)
	if nil != err {
		return nil, err
	}

	r, err := git.PlainOpen(path)
	if nil != err {
		return nil, err
	}

	w, err := r.Worktree()
	if nil != err {
		return nil, err
	}

	// ... we need a file to commit so let's create a new file inside of the
	// worktree of the project using the go standard library.
	Info("echo \"hello world!\" > example-git-file")
	filename := filepath.Join(workFolder, "example-git-file")
	err = ioutil.WriteFile(filename, []byte("hello world!"), 0644)
	if nil != err {
		return nil, err
	}

	// Adds the new file to the staging area.
	Info("git add example-git-file")
	_, err = w.Add("example-git-file")
	if nil != err {
		return nil, err
	}

	// We can verify the current status of the worktree using the method Status.
	Info("git status --porcelain")
	status, err := w.Status()
	if nil != err {
		return nil, err
	}

	fmt.Println(status)

	// Commits the current staging area to the repository, with the new file
	// just created. We should provide the object.Signature of Author of the
	// commit.
	Info("git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Steven Yang",
			Email: "chuyang@tibco.com",
			When:  time.Now(),
		},
	})

	CheckIfError(err)

	// Prints the current HEAD to verify that all worked well.
	Info("git show -s")
	obj, err := r.CommitObject(commit)
	if nil != err {
		return nil, err
	}

	fmt.Println(obj)
	return map[string]interface{}{}, nil
}

func (a *GitHubAccessorActivity) push(url string, project string, workFolder string, username string, password string) (map[string]interface{}, error) {
	//CheckArgs("<repository-path>")

	r, err := git.PlainOpen(workFolder)
	if nil != err {
		return nil, err
	}

	Info("git push")
	// push using default options
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
	})
	return map[string]interface{}{}, err
}

func ActivityId(ctx activity.Context) string {
	return fmt.Sprintf("%s_%s", ctx.FlowDetails().Name(), ctx.TaskName())
}

func readFile(filename string) (string, error) {
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("File reading error", err)
		return "", err
	}
	//fmt.Println("Contents of file:", string(fileContent))
	return string(fileContent), nil
}
