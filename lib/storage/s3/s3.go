package s3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	glob "github.com/ganbarodigital/go_glob"
	"github.com/getevo/evo/v2/lib/storage/lib"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type Driver struct {
	Dir      string
	name     string
	bucket   string
	key      string
	secret   string
	endpoint string
	region   string
	token    string
	session  *s3.Client
}

func (driver *Driver) SetName(name string) {
	driver.name = name
}

func (driver *Driver) Name() string {
	return driver.name
}

func (driver *Driver) Type() string {
	return "s3"
}

func (driver *Driver) Remove(path string) error {
	path = driver.getRealPath(path)
	_, err := driver.session.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (driver *Driver) RemoveAll(path string) error {
	path = driver.getRealPath(path)

	var marker *string
	for {
		res, err := driver.session.ListObjects(
			context.Background(),
			&s3.ListObjectsInput{
				Bucket: aws.String(driver.bucket),
				Marker: marker,
				Prefix: aws.String(path), // e.g. Must end with a "/" for a directory
			},
		)
		if err != nil {
			return err
		}

		for _, obj := range res.Contents {
			_, err := driver.session.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
				Bucket: aws.String(driver.bucket),
				Key:    obj.Key,
			})
			if err != nil {
				return err
			}
		}

		marker = res.NextMarker
		if marker == nil {
			break
		}
	}

	return nil
}

func (driver *Driver) List(path string, recursive ...bool) ([]lib.FileInfo, error) {
	path = driver.getRealPath(path)
	var p = strings.Replace(path, `/`, `\`, -1)
	var result []lib.FileInfo
	var marker *string
	var m = map[string]bool{}
	for {
		res, err := driver.session.ListObjects(
			context.Background(),
			&s3.ListObjectsInput{
				Bucket: aws.String(driver.bucket),
				Marker: marker,
				Prefix: aws.String(path), // e.g. Must end with a "/" for a directory
			},
		)
		if err != nil {
			return nil, err
		}

		for _, obj := range res.Contents {
			if strings.HasSuffix(*obj.Key, "/.ignore") {
				continue
			}
			if len(recursive) == 0 || !recursive[0] {
				if filepath.Dir(*obj.Key) == p {
					var info = lib.NewFileInfo(*obj.Key, obj.Size, 0644, *obj.LastModified, false, nil, driver)
					result = append(result, info)
				} else {
					var first = strings.Split(strings.Trim((*obj.Key)[len(path):], "/"), "/")[0]
					if ok, _ := m[first]; !ok {
						m[first] = true
						var info = lib.NewFileInfo(path+"/"+first, 0, 0755, *obj.LastModified, true, nil, driver)
						result = append(result, info)
					}
				}
			} else {
				var info = lib.NewFileInfo(*obj.Key, obj.Size, 0644, *obj.LastModified, false, nil, driver)
				var chunks = strings.Split(*obj.Key, "/")
				var build = ""

				for i, item := range chunks {
					if i == 0 {
						continue
					}
					build += "/" + item
					if ok, _ := m[build]; !ok {
						m[build] = true
						var dirInfo = lib.NewFileInfo(strings.TrimLeft(build, "/"), 0, 0755, *obj.LastModified, true, nil, driver)
						result = append(result, dirInfo)

					}

				}

				result = append(result, info)
			}
		}

		marker = res.NextMarker
		if marker == nil {
			break
		}
	}
	return result, nil
}

func (driver *Driver) Search(match string) ([]lib.FileInfo, error) {
	var parts = strings.Split(match, "*")
	var path = driver.getRealPath(parts[0])
	var globPath = driver.getRealPath(match)
	g := glob.NewGlob(globPath)
	var result []lib.FileInfo
	var marker *string
	var m = map[string]bool{}
	for {
		res, err := driver.session.ListObjects(
			context.Background(),
			&s3.ListObjectsInput{
				Bucket: aws.String(driver.bucket),
				Marker: marker,
				Prefix: aws.String(path), // e.g. Must end with a "/" for a directory
			},
		)
		if err != nil {
			return nil, err
		}

		for _, obj := range res.Contents {

			if ok, _ := g.Match(*obj.Key); ok {
				var info = lib.NewFileInfo(*obj.Key, obj.Size, 0644, *obj.LastModified, false, nil, driver)
				var chunks = strings.Split(*obj.Key, "/")
				var build = ""

				for i, item := range chunks {
					if i == 0 {
						continue
					}
					build += "/" + item
					if ok, _ := m[build]; !ok {
						m[build] = true
						var dirInfo = lib.NewFileInfo(strings.TrimLeft(build, "/"), 0, 0755, *obj.LastModified, true, nil, driver)
						result = append(result, dirInfo)

					}

				}
				result = append(result, info)
			}

		}

		marker = res.NextMarker
		if marker == nil {
			break
		}
	}
	return result, nil

}

var settings = lib.StorageSettings(`^(?P<proto>s3):\/\/(?P<keyid>[\S\s]+)\:(?P<secret>[\S\s]+)@(?P<endpoint>[a-zA-Z0-9\-\_\.\:]+)\/(?P<bucket>[a-zA-Z0-9\-\_\.]+)(?P<dir>.*)`)

func (driver *Driver) Init(input string) error {
	config, err := settings.Parse(input)
	if err != nil {
		return err
	}
	driver.secret = config["secret"]
	driver.key = config["keyid"]
	driver.endpoint = config["endpoint"]
	driver.Dir = strings.TrimLeft(config["dir"], "/")
	driver.bucket = config["bucket"]

	if v, ok := config["region"]; ok {
		config["region"] = v
	} else {
		config["region"] = "us-east-1"
	}
	driver.region = config["region"]
	var session = s3.NewFromConfig(aws.Config{
		Region:      driver.region,
		Credentials: credentials.NewStaticCredentialsProvider(driver.key, driver.secret, driver.token),
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               "https://" + driver.endpoint,
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		}),
	}, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	driver.session = session
	return nil
}

func (driver *Driver) SetWorkingDir(path string) error {
	driver.Dir = path
	return nil
}

func (driver *Driver) WorkingDir() string {
	return driver.Dir
}

func (driver *Driver) Touch(path string) error {
	driver.Write(path, "")
	return nil
}

func (driver *Driver) WriteJson(path string, content any) error {
	var b, err = json.Marshal(content)
	if err != nil {
		return err
	}
	return driver.Write(path, b)
}

func (driver *Driver) Append(path string, content any) error {
	path = driver.getRealPath(path)
	b, err := driver.ReadAll(path)
	if err != nil {
		return err
	}

	var data *bytes.Reader
	switch v := content.(type) {
	case string:
		data = bytes.NewReader(append(b, []byte(v)...))
	case []byte:
		data = bytes.NewReader(append(b, v...))
	default:
		return fmt.Errorf("invalid content type")
	}

	_, err = driver.session.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
		Body:   data,
	})

	return err
}

func (driver *Driver) SetMetadata(path string, meta lib.Metadata) error {
	path = driver.getRealPath(path)
	var tmp = map[string]string{}
	for key, _ := range meta {
		if key == "is_dir" || key == "size" || key == "name" || key == "mod_time" || key == "mod" {
			continue
		}
		tmp[key] = fmt.Sprint(meta[key])
	}
	var _, err = driver.session.CopyObject(context.Background(), &s3.CopyObjectInput{
		Key:               aws.String(path),
		CopySource:        aws.String(driver.bucket + "/" + path),
		Bucket:            aws.String(driver.bucket),
		MetadataDirective: types.MetadataDirectiveReplace,
		Metadata:          tmp,
	})
	return err
}

func (driver *Driver) GetMetadata(path string) (*lib.Metadata, error) {
	path = driver.getRealPath(path)
	object, err := driver.session.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	var meta = lib.Metadata(object.Metadata)
	meta["name"] = filepath.Base(path)
	meta["size"] = fmt.Sprint(object.ContentLength)
	meta["is_dir"] = "false"
	meta["mode"] = "0644"
	meta["mod_time"] = object.LastModified.String()

	return &meta, nil
}

func (driver *Driver) CopyDir(src, dest string) error {
	src = driver.getRealPath(src)
	dest = driver.getRealPath(dest)

	var marker *string
	for {
		res, err := driver.session.ListObjects(
			context.Background(),
			&s3.ListObjectsInput{
				Bucket: aws.String(driver.bucket),
				Marker: marker,
				Prefix: aws.String(src), // e.g. Must end with a "/" for a directory
			},
		)
		if err != nil {
			return err
		}

		for _, obj := range res.Contents {
			var file = (*obj.Key)[len(src):]

			var _, err = driver.session.CopyObject(context.Background(), &s3.CopyObjectInput{
				Key:               aws.String(dest + file),
				CopySource:        aws.String(driver.bucket + "/" + src + file),
				Bucket:            aws.String(driver.bucket),
				MetadataDirective: types.MetadataDirectiveCopy,
			})
			if err != nil {
				return err
			}
		}

		marker = res.NextMarker
		if marker == nil {
			break
		}
	}

	return nil
}

func (driver *Driver) CopyFile(src, dest string) error {
	src = driver.getRealPath(src)
	dest = driver.getRealPath(dest)

	var _, err = driver.session.CopyObject(context.Background(), &s3.CopyObjectInput{
		Key:               aws.String(dest),
		CopySource:        aws.String(driver.bucket + "/" + src),
		Bucket:            aws.String(driver.bucket),
		MetadataDirective: types.MetadataDirectiveCopy,
	})
	return err
}

func (driver *Driver) Stat(path string) (lib.FileInfo, error) {
	path = driver.getRealPath(path)
	object, err := driver.session.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return lib.FileInfo{}, err
	}
	return lib.NewFileInfo(path, object.ContentLength, 0644, *object.LastModified, false, nil, driver), nil

}

func (driver *Driver) IsFileExists(path string) bool {
	path = driver.getRealPath(path)
	_, err := driver.session.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
	})
	return err == nil
}

func (driver *Driver) IsDirExists(path string) bool {
	path = driver.getRealPath(path)

	var result, err = driver.session.ListObjects(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(driver.bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return false
	}
	if len(result.Contents) == 0 {
		return false
	}
	for _, item := range result.Contents {
		if *item.Key == path {
			return false
		}
	}

	return true
}

func (driver *Driver) IsDir(path string) bool {
	return driver.IsDirExists(path)
}

func (driver *Driver) ReadAll(path string) ([]byte, error) {
	path = driver.getRealPath(path)

	result, err := driver.session.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	var reader = bufio.NewReader(result.Body)
	body := bytes.NewBuffer(nil)
	io.Copy(body, reader)
	return body.Bytes(), nil
}

func (driver *Driver) ReadAllString(path string) (string, error) {
	var content, err = driver.ReadAll(path)
	return string(content), err
}

func (driver *Driver) Mkdir(path string, perm ...fs.FileMode) error {
	return driver.Write(path+"/.ignore", "")
}

func (driver *Driver) MkdirAll(path string, perm ...fs.FileMode) error {
	return driver.Write(path+"/.ignore", "")
}

func (driver *Driver) Write(path string, content any) error {
	path = driver.getRealPath(path)

	var data *bytes.Reader
	switch v := content.(type) {
	case string:
		data = bytes.NewReader([]byte(v))
	case []byte:
		data = bytes.NewReader(v)
	default:
		return fmt.Errorf("invalid content type")
	}

	_, err := driver.session.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(driver.bucket),
		Key:    aws.String(path),
		Body:   data,
	})

	return err

}

func (driver *Driver) getRealPath(path string) string {
	path = filepath.Clean(filepath.Join(driver.Dir, path))
	path = strings.Replace(path, `\`, `/`, -1)
	return path
}
