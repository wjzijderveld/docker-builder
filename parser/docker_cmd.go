package parser

import (
	"io"
	"os/exec"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/modcloth/go-fileutils"

	"github.com/rafecolton/docker-builder/dclient"
)

/*
DockerCmdOpts is an options struct for the options required by the various
structs that implement the DockerCmd interface
*/
type DockerCmdOpts struct {
	DockerClient dclient.DockerClient
	Image        string
	Workdir      string
	Stdout       io.Writer
	Stderr       io.Writer
	SkipPush     bool
	ImageUUID    string
}

/*
DockerCmd is an interface that wraps the various docker command types.
*/
type DockerCmd interface {
	// Run() runs the underlying command. The string return value is expected
	// to be the ID of the image being operated on
	Run() (string, error)

	// Message() returns a string representation of the command if it were to
	// be run on the command line
	Message() string

	// WithOpts sets the options for the command. It is expected to return the
	// same DockerCmd in a state in which the Run() function can be called
	// immediately after without error (i.e.`dockerCmdInstance.WithOpts(opts).Run()`)
	WithOpts(opts *DockerCmdOpts) DockerCmd
}

//BuildCmd is a wrapper for the os/exec call for `docker build`
type BuildCmd struct {
	Cmd  *exec.Cmd
	opts *DockerCmdOpts
}

//WithOpts sets options required for the BuildCmd
func (b *BuildCmd) WithOpts(opts *DockerCmdOpts) DockerCmd {
	b.opts = opts
	return b
}

//Run is the command that actually calls docker build shell command.  Determine
//the image ID for the resulting image and return that as well.
func (b *BuildCmd) Run() (string, error) {
	cmd := b.Cmd
	opts := b.opts

	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	cmd.Dir = opts.Workdir

	dockerExePath, err := fileutils.Which("docker")
	if err != nil {
		return "", err
	}

	cmd.Path = dockerExePath

	if err := b.Cmd.Run(); err != nil {
		return "", err
	}

	imageID, err := opts.DockerClient.LatestImageTaggedWithUUID(opts.ImageUUID)
	if err != nil {
		return "", err
	}

	return imageID, nil
}

//Message returns the shell command that gets run for docker build commands
func (b *BuildCmd) Message() string {
	return strings.Join(b.Cmd.Args, " ")
}

//TagCmd is a wrapper for the docker TagImage functionality
type TagCmd struct {
	TagFunc func(name string, opts docker.TagImageOptions) error
	Image   string
	Force   bool
	Tag     string
	Repo    string
	msg     string
}

//WithOpts sets options required for the TagCmd
func (t *TagCmd) WithOpts(opts *DockerCmdOpts) DockerCmd {
	t.Image = opts.Image
	t.TagFunc = opts.DockerClient.TagImage
	return t
}

//Run is the command that actually calls TagImage to do the tagging
func (t *TagCmd) Run() (string, error) {
	var opts = &docker.TagImageOptions{
		Force: t.Force,
		Repo:  t.Repo,
		Tag:   t.Tag,
	}
	return t.Image, t.TagFunc(t.Image, *opts)
}

//Message returns the shell command that would be equivalent to the TagImage command
func (t *TagCmd) Message() string {
	if t.msg == "" {
		msg := []string{"docker", "tag"}
		if t.Force {
			msg = append(msg, "--force")
		}
		msg = append(msg, t.Image)
		msg = append(msg, t.Repo+":"+t.Tag)
		t.msg = strings.Join(msg, " ")
	}

	return t.msg
}

//PushCmd is a wrapper for the docker PushImage functionality
type PushCmd struct {
	PushFunc     func(opts docker.PushImageOptions, auth docker.AuthConfiguration) error
	Image        string
	Tag          string
	Registry     string
	AuthUn       string
	AuthPwd      string
	AuthEmail    string
	OutputStream io.Writer

	skip    bool
	imageID string
}

//WithOpts sets options required for the PushCmd
func (p *PushCmd) WithOpts(opts *DockerCmdOpts) DockerCmd {
	p.OutputStream = opts.Stdout
	p.PushFunc = opts.DockerClient.PushImage
	p.skip = opts.SkipPush
	p.imageID = opts.Image
	return p
}

//Run is the command that actually calls PushImage to do the pushing
func (p *PushCmd) Run() (string, error) {
	if p.skip {
		return p.imageID, nil
	}

	auth := &docker.AuthConfiguration{
		Username: p.AuthUn,
		Password: p.AuthPwd,
		Email:    p.AuthEmail,
	}
	opts := &docker.PushImageOptions{
		Name:         p.Image,
		Tag:          p.Tag,
		Registry:     p.Registry,
		OutputStream: p.OutputStream,
	}
	return p.imageID, p.PushFunc(*opts, *auth)
}

//Message returns the shell command that would be equivalent to the PushImage command
func (p *PushCmd) Message() string {
	return "docker push " + p.Image + ":" + p.Tag
}
