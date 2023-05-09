package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var mu sync.Mutex

// Config represents SSH connection parameters.
type Config struct {
	Username     string
	Password     string
	PrivateKey   string
	Server       string
	KeyExchanges []string

	Timeout time.Duration
}

// Client provides basic functionality to interact with a SFTP server.
type Client struct {
	config     Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// New initialises SSH and SFTP clients and returns Client type to use.
func New(config Config) (*Client, error) {
	c := &Client{
		config: config,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	return c, nil
}

// Create creates a remote/destination file for I/O.
func (c *Client) Create(filePath string) (io.ReadWriteCloser, error) {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return c.sftpClient.Create(filePath)
}

// Upload writes local/source file data streams to remote/destination file.
func (c *Client) Upload(source io.Reader, destination io.Writer, size int) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	chunk := make([]byte, size)

	for {
		num, err := source.Read(chunk)
		if err == io.EOF {
			tot, err := destination.Write(chunk[:num])
			if err != nil {
				return err
			}

			if tot != len(chunk[:num]) {
				return fmt.Errorf("failed to write stream")
			}

			return nil
		}

		if err != nil {
			return err
		}

		tot, err := destination.Write(chunk[:num])
		if err != nil {
			return err
		}

		if tot != len(chunk[:num]) {
			return fmt.Errorf("failed to write stream")
		}
	}
}

// Download returns remote/destination file for reading.
func (c *Client) Download(filePath string) (io.ReadCloser, error) {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return c.sftpClient.Open(filePath)
}

// Info gets the details of a file. If the file was not found, an error is returned.
func (c *Client) Info(filePath string) (os.FileInfo, error) {
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	info, err := c.sftpClient.Lstat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file stats: %w", err)
	}

	return info, nil
}

// Close closes open connections.
func (c *Client) Close() {
	if c.sftpClient != nil {
		c.sftpClient.Close()
	}
	if c.sshClient != nil {
		c.sshClient.Close()
	}
}

// connect initialises a new SSH and SFTP client only if they were not
// initialised before at all and, they were initialised but the SSH
// connection was lost for any reason.
func (c *Client) connect() error {
	if c.sshClient != nil {
		_, _, err := c.sshClient.SendRequest("keepalive", false, nil)
		if err == nil {
			return nil
		}
	}

	auth := ssh.Password(c.config.Password)
	if c.config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.config.PrivateKey))
		if err != nil {
			return fmt.Errorf("ssh parse private key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	}

	cfg := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error { return nil },
		Timeout:         c.config.Timeout,
		Config: ssh.Config{
			KeyExchanges: c.config.KeyExchanges,
		},
	}

	sshClient, err := ssh.Dial("tcp", c.config.Server, cfg)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	c.sshClient = sshClient

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("sftp new client: %w", err)
	}
	c.sftpClient = sftpClient

	return nil
}

func (c *Client) Remove(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient.Remove(path)
}

func (c *Client) RemoveDir(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient.RemoveDirectory(path)
}

func (c *Client) Write(path string, data *bytes.Reader) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	dstFile, err := c.sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, data)
	return err
}

func (c *Client) Append(path string, data *bytes.Reader) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	dstFile, err := c.sftpClient.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, data)
	return err
}

func (c *Client) List(path string) ([]os.FileInfo, error) {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	files, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *Client) Mkdir(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient.Mkdir(path)
}

func (c *Client) MkdirAll(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient.MkdirAll(path)
}

func (c *Client) Stat(path string) (os.FileInfo, error) {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient.Stat(path)
}

func (c *Client) Connection() (*sftp.Client, error) {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return c.sftpClient, nil
}

func (c *Client) Touch(path string) error {
	mu.Lock()
	defer mu.Unlock()
	if err := c.connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	_, err := c.sftpClient.Stat(path)
	if err != nil {
		file, err := c.sftpClient.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = c.sftpClient.Chtimes(path, currentTime, currentTime)
		if err != nil {
			return err
		}
	}

	return nil
}
