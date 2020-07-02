package sftp

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strings"
	"time"
)

type UploadAction int

const (
	UPLOAD_DONT_OVERWRITE UploadAction = 0
	UPLOAD_OVERWRITE      UploadAction = 1
	UPLOAD_CONTINUE       UploadAction = 2
)

// SFTP sftp struct
type SFTP struct {
	Username        string
	Password        string
	Port            string
	Host            string
	HostKeyCallback ssh.HostKeyCallback
	AuthMethod      []ssh.AuthMethod
	Connected       bool
	Connection      *ssh.Client
	Client          *sftp.Client
	Timeout         time.Duration
}

// Connect connect to sftp server
func (s *SFTP) Connect() error {
	if s.HostKeyCallback == nil {
		s.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}
	if len(s.AuthMethod) == 0 && s.Password != "" {
		s.AuthMethod = []ssh.AuthMethod{
			ssh.Password(s.Password),
			ssh.Password(s.Password),
		}
	}

	if s.Timeout == 0 {
		s.Timeout = 3 * time.Second
	}

	config := &ssh.ClientConfig{
		User:            s.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            s.AuthMethod,
		Timeout:         s.Timeout,
	}
	config.SetDefaults()

	host := strings.Split(s.Host, ":")
	if len(host) == 2 {
		s.Host = host[0]
		s.Port = host[1]
	}
	if s.Port == "" {
		s.Port = "22"
	}

	conn, err := ssh.Dial("tcp", s.Host+":"+s.Port, config)
	if err != nil {
		return err
	}
	s.Connection = conn
	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	s.Client = client
	s.Connected = true
	return nil
}

func (s *SFTP) interactivePassword(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
	answers = make([]string, len(questions))
	for n := range questions {
		answers[n] = s.Password
	}
	return answers, nil
}

// ListDirectoryRecursive list directory info recursive
func (s *SFTP) ListDirectoryRecursive(dir string) ([]string, error) {
	var paths []string
	w := s.Client.Walk(dir)
	for w.Step() {
		if w.Err() != nil {
			return paths, w.Err()
		}

		paths = append(paths, w.Path())
	}
	return paths, nil
}

// Close close connection
func (s *SFTP) Close() {
	s.Client.Close()
	s.Connection.Close()
}

// ListDirectory list directory
func (s *SFTP) ListDirectory(dir string) ([]os.FileInfo, error) {
	list, err := s.Client.ReadDir(dir)
	if err != nil {
		return list, err
	}

	return list, nil
}

// Download download a file
func (s *SFTP) Download(remote, local string) (int64, error) {
	dstFile, err := os.Create(local)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	// open source file
	srcFile, err := s.Client.Open(remote)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()
	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return bytes, err
	}
	// flush in-memory copy
	err = dstFile.Sync()
	if err != nil {
		return bytes, err
	}

	return bytes, nil
}

// Stat return file info
func (s *SFTP) Stat(path string) (os.FileInfo, error) {
	return s.Client.Lstat(path)
}

// Delete deletes a file
func (s *SFTP) Delete(path string) error {
	stat, err := s.Client.Lstat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return s.Client.RemoveDirectory(path)
	}
	return s.Client.Remove(path)
}

// Exist check if file is exist
func (s *SFTP) Exist(path string) bool {
	_, err := s.Client.Lstat(path)
	return err == nil
}

// Upload upload a local file to sftp
func (s *SFTP) Upload(local, remote string, overwrite UploadAction) (int64, error) {
	remote_tmp := remote + "_tmp"
	var err error
	_, err = s.Stat(remote)
	if overwrite == UPLOAD_DONT_OVERWRITE && err != nil {
		return 0, fmt.Errorf("remote file exist")
	}
	if overwrite == UPLOAD_OVERWRITE && err == nil {
		err := s.Client.Remove(remote)
		s.Client.Remove(remote_tmp)
		if err != nil {
			return 0, err
		}
	}
	var dstFile *sftp.File
	dstFile, err = s.Client.Create(remote_tmp)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	// create source file
	srcFile, err := os.Open(local)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()
	// copy source file to destination file
	var bytes int64
	bytes, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return bytes, err
	}

	err = s.Client.Remove(remote)
	dstFile.Close()

	err = s.Client.Rename(remote_tmp, remote)
	if err != nil {
		return 0, err
	}
	s.Client.Remove(remote_tmp)
	return bytes, nil
}
