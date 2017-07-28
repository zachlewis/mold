package main

import (
	"fmt"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type gitVersion struct {
	r       *git.Repository
	visited map[plumbing.Hash]struct{}

	head      *plumbing.Reference
	tags      map[plumbing.Hash]*plumbing.Reference
	latestTag *plumbing.Reference
	distance  int
}

func newGitVersion(path string) (*gitVersion, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return &gitVersion{}, err
	}

	gt := &gitVersion{r: r}
	gt.initVersion()
	return gt, nil
}

func (gt *gitVersion) Commit() string {
	if gt.head == nil {
		return ""
	}
	return gt.head.String()[:7]
}

func (gt *gitVersion) TagVersion() string {
	if gt.latestTag == nil {
		return "0.0.0"
	}
	return strings.TrimPrefix(gt.latestTag.Name().Short(), "v")
}

func (gt *gitVersion) Version() string {
	cmt := gt.Commit()
	tv := gt.TagVersion()
	if cmt == "" || gt.distance == 0 {
		return tv
	}
	return fmt.Sprintf("%s-%d-%s", tv, gt.distance, cmt)
}

func (gt *gitVersion) initVersion() {
	var err error
	if gt.head, err = gt.r.Head(); err != nil {
		return
	}
	if gt.head.IsTag() {
		gt.latestTag = gt.head
		return
	}

	gt.getTags()
	gt.visited = map[plumbing.Hash]struct{}{}
	gt.recurse(gt.head.Hash())
}

func (gt *gitVersion) getTags() error {
	gt.tags = map[plumbing.Hash]*plumbing.Reference{}
	iter, err := gt.r.Tags()
	if err != nil {
		return err
	}

	iter.ForEach(func(arg1 *plumbing.Reference) error {

		if obj, err := gt.r.Object(plumbing.AnyObject, arg1.Hash()); err == nil {
			switch obj.(type) {
			case *object.Tag:
				tag := obj.(*object.Tag)
				if c, e := tag.Commit(); e == nil {
					gt.tags[c.Hash] = arg1
				}

			case *object.Commit:
				gt.tags[obj.ID()] = arg1
			}
		}

		return nil
	})
	return nil
}

func (gt *gitVersion) recurse(h plumbing.Hash) error {
	gt.visited[h] = struct{}{}

	obj, err := gt.r.Object(plumbing.AnyObject, h)
	if err != nil {
		return err
	}

	if obj.Type() != plumbing.CommitObject {
		return nil
	}
	cmt := obj.(*object.Commit)

	if tag, ok := gt.tags[cmt.Hash]; ok {
		//fmt.Println("FOUND", tag)
		gt.latestTag = tag
		return nil
	}

	gt.distance++

	iter := cmt.Parents()
	iter.ForEach(func(pc *object.Commit) error {

		if _, ok := gt.visited[pc.Hash]; !ok {
			gt.recurse(pc.Hash)
		}
		return nil
	})
	if tree, err := cmt.Tree(); err == nil {
		gt.recurse(tree.Hash)
	}

	return nil
}
